package _plugin

import (
	"encoding/base64"
	"errors"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	"gorm.io/gorm"
)

func EncryptBaiduIDData() error {
	if !share.IsPureGO {
		return errors.New("ERROR: Do not use encrypt/decrypt in compat mode")
	}

	offset := 0
	var baiduID []*model.TcBaiduid
	for {
		_function.GormDB.R.Model(&model.TcBaiduid{}).Select("id", "bduss", "stoken").Offset(offset).Limit(100).Find(&baiduID)
		if len(baiduID) <= 0 {
			break
		}

		err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			for _, baiduIDItem := range baiduID {
				encryptedBDUSS, _ := _function.AES256GCMEncrypt(baiduIDItem.Bduss, share.DataEncryptKeyByte)
				baiduIDItem.Bduss = strings.ReplaceAll(base64.URLEncoding.EncodeToString(encryptedBDUSS), "=", "")

				encryptedStoken, _ := _function.AES256GCMEncrypt(baiduIDItem.Stoken, share.DataEncryptKeyByte)
				baiduIDItem.Stoken = strings.ReplaceAll(base64.URLEncoding.EncodeToString(encryptedStoken), "=", "")

				if err := tx.Model(&model.TcBaiduid{}).Select("bduss", "stoken").Where("id = ?", baiduIDItem.ID).Updates(&baiduIDItem).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		offset += 100
	}
	return nil
}

func DecryptBaiduIDData() error {
	if !share.IsPureGO {
		return errors.New("ERROR: Do not use encrypt/decrypt in compat mode")
	}

	offset := 0
	var baiduID []*model.TcBaiduid
	for {
		_function.GormDB.R.Model(&model.TcBaiduid{}).Select("id", "bduss", "stoken").Offset(offset).Limit(100).Find(&baiduID)
		if len(baiduID) <= 0 {
			break
		}
		err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			for _, baiduIDItem := range baiduID {
				decryptedBDUSS, _ := _function.AES256GCMDecrypt(strings.ReplaceAll(baiduIDItem.Bduss, "=", ""), share.DataEncryptKeyByte)
				baiduIDItem.Bduss = string(decryptedBDUSS)

				decryptedStoken, _ := _function.AES256GCMDecrypt(strings.ReplaceAll(baiduIDItem.Stoken, "=", ""), share.DataEncryptKeyByte)
				baiduIDItem.Stoken = string(decryptedStoken)

				if err := tx.Model(&model.TcBaiduid{}).Select("bduss", "stoken").Where("id = ?", baiduIDItem.ID).Updates(&baiduIDItem).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		offset += 100
	}
	return nil
}
