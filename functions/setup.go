package _function

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/dao/model"
)

func SetupSystem(dbMode string, _tc_mysql string, _tc_sqlite string, _tc_init_system string) {
	fmt.Println("📌现在正在安装 TbSign➡️，如果数据库内含有数据，这样做会导致数据丢失，请提前做好备份。")
	fmt.Println("如果已经完成备份，请输入以下随机数字并按下回车（显示为 \"--> 1234 <--\" 代表需要输入 \"1234\"）")
	randValue := strconv.Itoa(int(rand.Int31()))
	fmt.Println("-->", randValue, "<--")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入: ")
	verifyText, _ := reader.ReadString('\n')
	verifyText = strings.TrimSpace(verifyText)
	if verifyText != randValue {
		fmt.Println("❌输入错误，请重试")
		os.Exit(0)
	}

	fmt.Println("⌛正在清理旧表")
	GormDB.W.Migrator().DropTable(&model.TcBaiduid{},
		&model.TcKdGrowth{},
		&model.TcOption{},
		&model.TcPlugin{},
		&model.TcTieba{},
		&model.TcUsersOption{},
		&model.TcUser{},
		&model.TcVer4BanList{},
		&model.TcVer4BanUserset{},
		&model.TcVer4RankLog{},
	)

	fmt.Println("⌛正在建立数据表和索引")
	if dbMode == "mysql" {
		for i, v := range strings.Split(_tc_mysql, ";") {
			if len(strings.TrimSpace(v)) == 0 {
				continue
			}
			fmt.Println("⌛导入第" + strconv.Itoa(i+1) + "项...")
			err := GormDB.W.Exec(v).Error
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		err := GormDB.W.Exec(_tc_sqlite).Error
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("⌛正在导入数据...")
	for i, v := range strings.Split(_tc_init_system, "\n") {
		fmt.Println("⌛导入第" + strconv.Itoa(i+1) + "项...")
		err := GormDB.W.Exec(v).Error
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("🔒注册管理员帐号...")
	fmt.Print("管理员用户名: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("❌无效用户名", err)
	}
	name = strings.TrimSuffix(name, "\n")
	if name == "" || strings.Contains(name, "@") {
		log.Fatal("❌无效用户名")
	}
	fmt.Print("管理员邮箱: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("❌无效邮箱", err)
	}
	email = strings.TrimSuffix(email, "\n")
	if !VerifyEmail(email) {
		log.Fatal("❌无效邮箱")
	}
	fmt.Print("管理员密码 (注意空格): ")
	password, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("❌无效密码", err)
	}
	password = strings.TrimSuffix(password, "\n")
	if password == "" {
		log.Fatal("❌无效密码")
	}

	passwordHash, err := CreatePasswordHash(password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("⌛正在注册管理员帐号...")
	GormDB.W.Create(&model.TcUser{
		ID:    1,
		Name:  name,
		Email: email,
		Pw:    string(passwordHash),
		Role:  "admin",
		T:     "tieba",
	})

	fmt.Println("🎉安装成功！请移除掉 `--setup=true` 后重新执行本文件以启动系统")
	os.Exit(0)
}
