package _api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
)

var PreCheckWhiteListWithoutFE = []string{
	"/",
	"/*",
	"/favicon.ico",
	"/robots.txt",
	"/passport/login",
	"/passport/signup",
	"/passport/reset/password",
	// "/tools/userinfo/tieba_uid/:tiebauid",
	// "/tools/userinfo/panel/:query_type/:user_value",
	// "/tools/tieba/fname_to_fid/:fname",
	"/config/page/login",
}
var PreCheckWhiteListWithFE = []string{
	"/api/passport/login",
	"/api/passport/signup",
	"/api/passport/reset/password",
	// "/api/tools/userinfo/tieba_uid/:tiebauid",
	// "/api/tools/userinfo/panel/:query_type/:user_value",
	// "/api/tools/tieba/fname_to_fid/:fname",
	"/api/config/page/login",
}

var RoleList = []string{"deleted", "banned", "user", "vip", "admin"}

func PreCheckWhiteListExists(path string) bool {
	if share.EnableFrontend {
		return slices.Contains(PreCheckWhiteListWithFE, path)
	} else {
		return slices.Contains(PreCheckWhiteListWithoutFE, path)
	}
}

func echoRobots(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /*")
}

func verifyAuthorization(authorization string) (string, string) {
	if authorization == "" {
		return "0", "guest"
	} else {
		// Parse the token

		_unverifiedToken, _, err := jwt.NewParser().ParseUnverified(authorization, &jwt.RegisteredClaims{})
		if err != nil {
			return "0", "guest"
		}

		var tmpUID string
		if claims, ok := _unverifiedToken.Claims.(*jwt.RegisteredClaims); ok {
			tmpUID = claims.Subject
		} else {
			return "0", "guest"
		}

		if key, ok := keyBucket.Load(tmpUID); ok {
			token, err := jwt.ParseWithClaims(authorization, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return &key.(*ecdsa.PrivateKey).PublicKey, nil
			}, jwt.WithIssuedAt(), jwt.WithIssuer("TbSign->"), jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"ES256"}))
			if err != nil || !token.Valid {
				return "0", "guest"
			} else {
				claims := token.Claims.(*jwt.RegisteredClaims)
				now := time.Now()
				//exp & nbf
				if now.After(claims.ExpiresAt.Time) || now.Before(claims.NotBefore.Time) {
					return "0", "guest"
				}

				uid, _ := claims.GetSubject()
				var accountInfo []*model.TcUser
				_function.GormDB.R.Where("id = ?", uid).Limit(1).Find(&accountInfo)
				if len(accountInfo) == 1 {
					return strconv.Itoa(int(accountInfo[0].ID)), accountInfo[0].Role
				} else {
					return "0", "guest"
				}
			}
		} else {
			return "0", "guest"
		}
	}
}

func sessionTokenBuilder(uid int32, password string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(int(uid)) + ":" + hex.EncodeToString(_function.GenHMAC256([]byte(password), []byte(strconv.Itoa(int(uid))+password)))))
}

var keyBucket sync.Map //= make(map[string]*ecdsa.PrivateKey) // uid -> key

func bearerTokenBuilder(uid string, forceUpdate bool) string {
	privateKey := new(ecdsa.PrivateKey)
	if key, ok := keyBucket.Load(uid); ok && !forceUpdate {
		privateKey = key.(*ecdsa.PrivateKey)
	} else {
		privateKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		keyBucket.Store(uid, privateKey)
	}

	// expire
	strCookieExpire := _function.GetOption("cktime")
	numberCookieExpire, err := strconv.ParseInt(strCookieExpire, 10, 64)
	if err != nil {
		numberCookieExpire = 10 * 60
	} else if numberCookieExpire > 10*24*60*60 {
		numberCookieExpire = 10 * 24 * 60 * 60
	} else if numberCookieExpire < 30 {
		numberCookieExpire = 30
	}

	// Create the Claims
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   uid,
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(numberCookieExpire) * time.Second)),
		Audience:  []string{"TbSign->"},
		Issuer:    "TbSign->",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	ss, _ := token.SignedString(privateKey)

	return ss
}
