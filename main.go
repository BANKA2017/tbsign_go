package main

import (
	"database/sql"
	_ "embed"
	"flag"
	"log"
	"os"
	"time"

	_api "github.com/BANKA2017/tbsign_go/api"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	_type "github.com/BANKA2017/tbsign_go/types"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

//go:embed assets/sql/tc_init_system.sql
var _tc_init_system string

//go:embed assets/sql/tc_mysql.sql
var _tc_mysql string

//go:embed assets/sql/tc_sqlite.sql
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
	logLevel := logger.Error
	if testMode {
		logLevel = logger.Info
	}

	if dbPath != "" {
		// sqlite
		dbMode = "sqlite"
		_function.GormDB.R, _function.GormDB.W, err = _function.ConnectToSQLite(dbPath, logLevel, "tbsign")
		if err != nil {
			log.Fatal("db:", err)
		}
	} else {
		// mysql
		if dbUsername == "" || dbPassword == "" {
			log.Fatal("global: Empty username or password")
		}
		dsn := dbUsername + ":" + dbPassword + "@tcp(" + dbEndpoint + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
		sqlDB, _ := sql.Open("mysql", dsn)
		defer sqlDB.Close()
		_function.GormDB.W, err = gorm.Open(mysql.New(mysql.Config{
			Conn: sqlDB,
		}), &gorm.Config{Logger: logger.Default.LogMode(logLevel)})
		_function.GormDB.R = _function.GormDB.W

		if err != nil {
			log.Fatal("db:", err)
		}
		log.Println("db: mysql connected!")
	}

	// setup
	if setup {
		_function.SetupSystem(dbMode, _tc_mysql, _tc_sqlite, _tc_init_system)
	}

	// init
	_function.InitClient()
	_function.GetOptionsAndPluginList()

	if enableApi {
		go _api.Api(address, "dbmode", dbMode, "testmode", testMode)
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
				for uid, value := range _function.ResetPwdList {
					if value.Expire < _function.Now.Unix() {
						delete(_function.ResetPwdList, uid)
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
