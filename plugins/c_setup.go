package _plugin

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/assets"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupSystem(dbMode, dbPath, dbUsername, dbPassword, dbEndpoint, dbName, dbTLSOption string, logLevel logger.LogLevel, dbExists, autoInstall bool, name, email, password string) {
	reader := bufio.NewReader(os.Stdin)
	var err error

	_tc_mysql, _ := assets.EmbeddedSQL.ReadFile("sql/tc_mysql.sql")
	_tc_sqlite, _ := assets.EmbeddedSQL.ReadFile("sql/tc_sqlite.sql")
	_tc_pgsql, _ := assets.EmbeddedSQL.ReadFile("sql/tc_pgsql.sql")

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

	// mysql
	if dbMode == "mysql" {
		if !dbExists {
			fmt.Println("⌛正在建立数据库:", dbName)
			err = _function.GormDB.W.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + ";").Error
			if err != nil {
				log.Fatal(err)
			} else {
				fmt.Println("已建立数据库:", dbName)
			}
		}
		_function.GormDB.R, _function.GormDB.W, err = _function.ConnectToMySQL(dbUsername, dbPassword, dbEndpoint, dbName, dbTLSOption, logLevel, "db")
		if err != nil {
			log.Fatal("db:", err)
		}
	}

	fmt.Println("⌛正在清理旧表")
	_function.GormDB.W.Migrator().DropTable(&model.TcBaiduid{},
		// &model.TcKdGrowth{},
		&model.TcOption{},
		&model.TcPlugin{},
		&model.TcTieba{},
		&model.TcUsersOption{},
		&model.TcUser{},
		// &model.TcVer4BanList{},
		// &model.TcVer4BanUserset{},
		// &model.TcVer4RankLog{},
	)

	fmt.Println("⌛正在建立数据表和索引")
	if dbMode == "mysql" {
		err = _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			for i, v := range strings.Split(string(_tc_mysql), ";") {
				if len(strings.TrimSpace(v)) == 0 {
					continue
				}
				fmt.Println("⌛导入第" + strconv.Itoa(i+1) + "项...")
				err := tx.Exec(v).Error
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	} else if dbMode == "pgsql" {
		err := _function.GormDB.W.Exec(string(_tc_pgsql)).Error
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := _function.GormDB.W.Exec(string(_tc_sqlite)).Error
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("⌛正在导入默认设置...")
	err = _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
		for key, value := range assets.DefaultOptions {
			if key == "go_ver" {
				value = "1"
			}
			fmt.Printf("%s: %s\n", key, value)
			err := _function.SetOption(key, value, tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println("⌛正在安装插件...")
	// for name, plugin := range PluginList {
	// 	fmt.Printf("⌛%s...\n", name)
	// 	err := plugin.Delete()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	err = plugin.Install()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

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
	if dbMode == "sqlite" {
		err := _function.GormDB.W.Exec("VACUUM;").Error
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("🎉安装成功！")
}
