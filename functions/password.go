package _function

import (
	"time"

	"github.com/BANKA2017/tbsign_go/model"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/crypto/bcrypt"
)

func CreatePasswordHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 12)
}

func VerifyPasswordHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

var PasswordCache = &KV[int, string]{
	KV: ttlcache.New[int, string](
		ttlcache.WithTTL[int, string](time.Hour),
		ttlcache.WithCapacity[int, string](100),
	),
}

func GetPassword(uid int) string {
	if pwd, ok := PasswordCache.Load(uid); ok {
		return pwd
	}
	var accountInfo []*model.TcUser
	GormDB.R.Select("pw").Where("id = ?", uid).Limit(1).Find(&accountInfo)

	if len(accountInfo) == 1 {
		PasswordCache.Store(uid, accountInfo[0].Pw, int64(ttlcache.DefaultTTL))
		return accountInfo[0].Pw
	}

	return ""
}

func UpdatePassword(uid int, newPassword string) (string, error) {
	hash, err := CreatePasswordHash(newPassword)
	if err != nil {
		return "", err
	}

	strPwd := string(hash)

	if err = GormDB.W.Model(&model.TcUser{}).Where("id = ?", uid).Update("pw", strPwd).Error; err != nil {
		return "", err
	}

	PasswordCache.Store(uid, strPwd, int64(ttlcache.DefaultTTL))
	return strPwd, nil
}
