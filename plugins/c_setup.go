package _plugin

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/assets"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/kdnetwork/code-snippet/go/db"
)

func SetupSystem(dbExists, autoInstall bool, name, email, password string) {
	reader := bufio.NewReader(os.Stdin)
	var err error

	fmt.Println("📌现在正在安装 TbSign➡️")
	if dbExists {
		fmt.Println("⚠️检测到数据库已存在，覆盖安装会导致原有数据丢失，请提前做好备份。")
	}
	if !autoInstall {
		fmt.Println("请输入以下随机数字并按下回车（显示为 \"--> 1234 <--\" 代表需要输入 \"1234\"）")
		randValue := strconv.Itoa(int(rand.Int31()))
		fmt.Println("-->", randValue, "<--")
		fmt.Print("请输入: ")
		verifyText, _ := reader.ReadString('\n')
		verifyText = strings.TrimSpace(verifyText)
		if verifyText != randValue {
			fmt.Println("❌输入错误，请重试")
			os.Exit(0)
		}
	}

	dbName := _function.GormDB.GetDB()

	if _function.GormDB.DBMode == db.DBModeMySQL || _function.GormDB.DBMode == db.DBModePostgreSQL {
		if !dbExists {
			if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(dbName) {
				log.Fatal("❌数据库名称无法用于 TbSign➡️")
			}

			fmt.Println("⌛正在建立数据库:", dbName)
			err = _function.GormDB.W.Exec(fmt.Sprintf("CREATE DATABASE `%s`;", dbName)).Error
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("已建立数据库:", dbName)
		}

		if err = _function.GormDB.Close(); err != nil {
			log.Fatal("db close:", err)
		}

		if err = _function.GormDB.Connect(); err != nil {
			log.Fatal("db:", err)
		}
	}

	fmt.Println("⌛正在清理旧表")
	_function.GormDB.W.Migrator().DropTable(&model.TcBaiduid{},
		&model.TcOption{},
		&model.TcPlugin{},
		&model.TcTieba{},
		&model.TcUsersOption{},
		&model.TcUser{},
	)

	fmt.Println("⌛正在建立数据表和索引")
	err = _function.GormDB.W.Migrator().CreateTable(
		&model.TcBaiduid{},
		&model.TcOption{},
		&model.TcPlugin{},
		&model.TcTieba{},
		&model.TcUsersOption{},
		&model.TcUser{},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("⌛正在导入默认设置...")

	assets.DefaultOptions["go_ver"] = "1"

	optionArray := make([]*model.TcOption, 0, len(assets.DefaultOptions))
	for k, v := range assets.DefaultOptions {
		optionArray = append(optionArray, &model.TcOption{Name: k, Value: v})
	}

	if err = _function.GormDB.W.Model(&model.TcOption{}).Create(optionArray).Error; err != nil {
		log.Fatal(err)
	}

	_function.InitOptions()

	fmt.Println("🔒注册管理员账号...")
	if !autoInstall {
		fmt.Print("管理员用户名: ")
		name, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal("❌无效用户名", err)
		}
		name = strings.TrimSpace(name)
		if name == "" || strings.Contains(name, "@") {
			log.Fatal("❌无效用户名")
		}
		fmt.Print("管理员邮箱: ")
		email, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal("❌无效邮箱", err)
		}
		email = strings.TrimSpace(email)
		if !_function.VerifyEmail(email) {
			log.Fatal("❌无效邮箱")
		}
		fmt.Print("管理员密码 (自动清理空格): ")
		password, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal("❌无效密码", err)
		}
		password = strings.TrimSpace(password)
		if password == "" {
			log.Fatal("❌无效密码")
		}
	} else {
		fmt.Println("管理员用户名:", name)
		fmt.Println("管理员邮箱:", email)
		fmt.Println("管理员密码:", password)
	}

	passwordHash, err := _function.CreatePasswordHash(password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("⌛正在注册管理员账号...")
	_function.GormDB.W.Create(&model.TcUser{
		ID:    1,
		Name:  name,
		Email: email,
		Pw:    string(passwordHash),
		Role:  "admin",
		T:     "tieba",
	})
	if _function.GormDB.DBMode == db.DBModeSQLite {
		err := _function.GormDB.W.Exec("VACUUM;").Error
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("🎉安装成功！")
}
