package main

import (
	"bufio"
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	_api "github.com/BANKA2017/tbsign_go/api"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	_type "github.com/BANKA2017/tbsign_go/types"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var dbUsername string
var dbPassword string
var dbEndpoint string
var dbName string

var dbPath string

var testMode bool
var enableApi bool

var address string

var setup bool

//go:embed assets/tc_init_system.sql
var _tc_init_system string

//go:embed assets/tc_mysql.sql
var _tc_mysql string

//go:embed assets/tc_sqlite.sql
var _tc_sqlite string

var err error

func main() {
	// sqlite
	flag.StringVar(&dbPath, "db_path", "", "Database path")

	// mysql
	flag.StringVar(&dbUsername, "username", "", "Username")
	flag.StringVar(&dbPassword, "pwd", "", "Password")
	flag.StringVar(&dbEndpoint, "endpoint", "127.0.0.1:3306", "endpoint")
	flag.StringVar(&dbName, "db", "tbsign", "Database name")

	//proxy
	flag.BoolVar(&_function.IgnoreProxy, "no_proxy", false, "Ignore the http proxy config from environment vars")

	// api
	flag.BoolVar(&enableApi, "api", false, "active backend endpoints")
	flag.StringVar(&address, "address", ":1323", "address :1323")

	// others
	flag.BoolVar(&testMode, "test", false, "Not send any requests to tieba servers")
	flag.BoolVar(&setup, "setup", false, "Init the system")

	flag.Parse()

	// from env
	if dbUsername == "" {
		dbUsername = os.Getenv("tc_username")
	}
	if dbPassword == "" {
		dbPassword = os.Getenv("tc_pwd")
	}
	if dbEndpoint == "" && os.Getenv("tc_endpoint") != "" {
		dbEndpoint = os.Getenv("tc_endpoint")
	}
	if dbName == "" && os.Getenv("tc_db") != "" {
		dbName = os.Getenv("tc_db")
	}
	if dbPath == "" && os.Getenv("tc_db_path") != "" {
		dbPath = os.Getenv("tc_db_path")
	}
	if !testMode && os.Getenv("tc_test") != "" {
		testMode = os.Getenv("tc_test") == "true"
	}
	if !enableApi && os.Getenv("tc_api") != "" {
		enableApi = os.Getenv("tc_api") == "true"
	}
	if address == ":1323" && os.Getenv("tc_address") != "" {
		address = os.Getenv("tc_address")
	}

	// connect to db
	dbMode := "mysql"

	if dbPath != "" {
		// sqlite
		_function.GormDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

		if err != nil {
			log.Fatal("db:", err)
		}
		_function.GormDB.Exec("PRAGMA journal_mode = WAL;PRAGMA busy_timeout = 5000;PRAGMA synchronous = NORMAL;PRAGMA cache_size = 100000;PRAGMA foreign_keys = true;PRAGMA temp_store = memory;")

		dbMode = "sqlite"
		log.Println("db: sqlite connected!")
	} else {
		// mysql
		if dbUsername == "" || dbPassword == "" {
			log.Fatal("global: Empty username or password")
		}
		dsn := dbUsername + ":" + dbPassword + "@tcp(" + dbEndpoint + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
		sqlDB, _ := sql.Open("mysql", dsn)
		_function.GormDB, err = gorm.Open(mysql.New(mysql.Config{
			Conn: sqlDB,
		}), &gorm.Config{})

		if err != nil {
			log.Fatal("db:", err)
		}
		log.Println("db: mysql connected!")
	}

	// setup
	if setup {
		fmt.Println("现在正在安装 TbSign➡️，如果数据库内含有数据，这样做会导致数据丢失，请提前做好备份，如果已经完成备份，请输入以下随机文字并按下回车（显示为 \"--> 1234 <--\" 代表需要输入 \"1234\"）")
		randValue := strconv.Itoa(int(rand.Int63()))
		fmt.Println("-->", randValue, "<--")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("请输入: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text != randValue {
			fmt.Println("输入错误，请重试")
			os.Exit(0)
		}

		fmt.Println("正在建立数据表和索引")
		if dbMode == "mysql" {
			err := _function.GormDB.Exec(_tc_mysql).Error
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err := _function.GormDB.Exec(_tc_sqlite).Error
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Println("正在导入数据")
		err := _function.GormDB.Exec(_tc_init_system).Error
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("🎉 安装成功！请移除掉 `--setup=true` 后重新执行本文件以启动系统")
		fmt.Println("🔔 首位注册的帐号将会被自动提权为管理员")
		os.Exit(0)
	}

	// init
	_function.InitClient()
	_function.GetOptionsAndPluginList()

	if enableApi {
		go _api.Api(address, "dbmode", dbMode, "testmode", testMode, "compat", _function.GetOption("core_version"))
	}

	// Interval
	oneSecondInterval := time.NewTicker(time.Second)
	defer oneSecondInterval.Stop()
	oneMinuteInterval := time.NewTicker(time.Minute)
	defer oneMinuteInterval.Stop()
	fourHoursInterval := time.NewTicker(time.Hour * 4)
	defer fourHoursInterval.Stop()

	// cron
	for {
		select {
		case <-oneSecondInterval.C:
			_function.UpdateNow()
		case <-oneMinuteInterval.C:
			if testMode {
				continue
			}
			_function.GetOptionsAndPluginList()
			_plugin.DoSignAction()
			_plugin.DoReSignAction()

			// plugins
			if p, ok := _function.PluginList["ver4_rank"]; ok && p.Status {
				go _plugin.DoForumSupportAction()
			}

			if p, ok := _function.PluginList["ver4_ban"]; ok && p.Status {
				go _plugin.LoopBanAction()
			}

			if p, ok := _function.PluginList["kd_growth"]; ok && p.Status {
				go _plugin.DoGrowthTasksAction()
			}

			// clean pwd list
			if len(_function.ResetPwdList) > 0 {
				for email, value := range _function.ResetPwdList {
					if value.Expire < _function.Now.Unix() {
						delete(_function.ResetPwdList, email)
					}
				}
			}
		case <-fourHoursInterval.C:
			if testMode {
				continue
			}
			_function.GetOptionsAndPluginList()
			if p, ok := _function.PluginList["ver4_ref"]; ok && p.Status {
				go _plugin.RefreshTiebaListAction()
			}

			// clean cookie/fid cache
			_function.CookieList = make(map[int32]_type.TypeCookie)
			_function.FidList = make(map[string]int64)
		}
	}
}
