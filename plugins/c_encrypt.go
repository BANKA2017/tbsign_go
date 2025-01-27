package _plugin

import (
	"errors"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	"gorm.io/gorm"
)

func EncryptBaiduIDData() error {
	if !share.IsPureGO {
		return errors.New("ERROR: 请不要在兼容模式下加密/解密数据")
	}

	if share.IsEncrypt {
		return errors.New("ERROR: 数据已经加密，请勿重复加密")
	}

	offset := 0

	// baidu user
	var baiduID []*model.TcBaiduid
	for {
		_function.GormDB.R.Model(&model.TcBaiduid{}).Select("id", "bduss", "stoken").Offset(offset).Limit(100).Find(&baiduID)
		if len(baiduID) <= 0 {
			break
		}

		err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			for _, baiduIDItem := range baiduID {
				encryptedBDUSS, _ := _function.AES256GCMEncrypt(baiduIDItem.Bduss, share.DataEncryptKeyByte)
				baiduIDItem.Bduss = _function.Base64URLEncode(encryptedBDUSS)

				encryptedStoken, _ := _function.AES256GCMEncrypt(baiduIDItem.Stoken, share.DataEncryptKeyByte)
				baiduIDItem.Stoken = _function.Base64URLEncode(encryptedStoken)

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

	// user options
	offset = 0
	var userOptions []*model.TcUsersOption
	for {
		_function.GormDB.R.Model(&model.TcUsersOption{}).Select("uid", "name", "value").Where("name IN (?)", _function.CanEncryptUserOption).Offset(offset).Limit(100).Find(&userOptions)
		if len(userOptions) <= 0 {
			break
		}

		err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			for _, userOption := range userOptions {
				encryptedUserOptionValue, _ := _function.AES256GCMEncrypt(userOption.Value, share.DataEncryptKeyByte)
				userOption.Value = _function.Base64URLEncode(encryptedUserOptionValue)

				if err := tx.Model(&model.TcUsersOption{}).Where("uid = ? AND name = ?", userOption.UID, userOption.Name).Update("value", userOption.Value).Error; err != nil {
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
	return _function.SetOption("go_encrypt", strconv.Itoa(int(time.Now().Unix())))
}

func DecryptBaiduIDData() error {
	if !share.IsPureGO {
		return errors.New("ERROR: 请不要在兼容模式下加密/解密数据")
	}

	if !share.IsEncrypt {
		return errors.New("ERROR: 已经是明文数据")
	}

	offset := 0

	// baidu user
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

	// user options
	offset = 0
	var userOptions []*model.TcUsersOption
	for {
		_function.GormDB.R.Model(&model.TcUsersOption{}).Select("uid", "name", "value").Where("name IN (?)", _function.CanEncryptUserOption).Offset(offset).Limit(100).Find(&userOptions)
		if len(userOptions) <= 0 {
			break
		}

		err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			for _, userOption := range userOptions {
				decryptedUserOptionValue, _ := _function.AES256GCMDecrypt(userOption.Value, share.DataEncryptKeyByte)
				userOption.Value = string(decryptedUserOptionValue)

				if err := tx.Model(&model.TcUsersOption{}).Where("uid = ? AND name = ?", userOption.UID, userOption.Name).Update("value", userOption.Value).Error; err != nil {
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

	return _function.SetOption("go_encrypt", "0")
}
