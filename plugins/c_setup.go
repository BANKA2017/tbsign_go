package _plugin

import (
	"bufio"
	"fmt"
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

func SetupSystem(dbExists, autoInstall bool, name, email, password string) bool {
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
			return false
		}
	}

	dbName := _function.GormDB.GetDB()

	if _function.GormDB.DBMode == db.DBModeMySQL || _function.GormDB.DBMode == db.DBModePostgreSQL {
		if !dbExists {
			if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(dbName) {
				fmt.Println("❌数据库名称无法用于 TbSign➡️")
				return false
			}

			fmt.Println("⌛正在建立数据库:", dbName)
			err = _function.GormDB.W.Exec(fmt.Sprintf("CREATE DATABASE `%s`;", dbName)).Error
			if err != nil {
				fmt.Println("❌创建数据库失败", err)
				return false
			}

			fmt.Println("已建立数据库:", dbName)
		}

		if err = _function.GormDB.Close(); err != nil {
			fmt.Println("❌关闭数据库失败", err)
			return false
		}

		if err = _function.GormDB.Connect(); err != nil {
			fmt.Println("❌连接数据库失败", err)
			return false
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
		fmt.Println("❌创建数据表失败", err)
		return false
	}

	fmt.Println("⌛正在导入默认设置...")

	assets.DefaultOptions["go_ver"] = "1"

	optionArray := make([]*model.TcOption, 0, len(assets.DefaultOptions))
	for k, v := range assets.DefaultOptions {
		optionArray = append(optionArray, &model.TcOption{Name: k, Value: v})
	}

	if err = _function.GormDB.W.Model(&model.TcOption{}).Create(optionArray).Error; err != nil {
		fmt.Println("❌导入默认设置失败", err)
		return false
	}

	_function.InitOptions()

	fmt.Println("🔒注册管理员账号...")
	if !autoInstall {
		fmt.Print("管理员用户名: ")
		name, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("❌无效用户名", err)
			return false
		}
		name = strings.TrimSpace(name)
		if name == "" || strings.Contains(name, "@") {
			fmt.Println("❌无效用户名")
			return false
		}
		fmt.Print("管理员邮箱: ")
		email, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("❌无效邮箱", err)
			return false
		}
		email = strings.TrimSpace(email)
		if !_function.VerifyEmail(email) {
			fmt.Println("❌无效邮箱")
			return false
		}
		fmt.Print("管理员密码 (自动清理空格): ")
		password, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("❌无效密码", err)
			return false
		}
		password = strings.TrimSpace(password)
		if password == "" {
			fmt.Println("❌无效密码")
			return false
		}
	} else {
		fmt.Println("管理员用户名:", name)
		fmt.Println("管理员邮箱:", email)
		fmt.Println("管理员密码:", strings.Repeat("*", len(password)))
	}

	passwordHash, err := _function.CreatePasswordHash(password)
	if err != nil {
		fmt.Println("❌保存密码失败", err)
		return false
	}

	fmt.Println("⌛正在注册管理员账号...")
	_function.GormDB.W.Create(&model.TcUser{
		ID:    1,
		Name:  name,
		Email: email,
		Pw:    string(passwordHash),
		Role:  _function.RoleAdmin,
		T:     "tieba",
	})
	if _function.GormDB.DBMode == db.DBModeSQLite {
		err := _function.GormDB.W.Exec("VACUUM;").Error
		if err != nil {
			fmt.Println("❌清理数据库失败", err)
			return false
		}
	}

	fmt.Println("✅安装成功")
	return true
}
