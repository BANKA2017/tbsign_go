package main

import (
	_ "embed"
	"flag"
	"log"
	"os"
	"time"
	_ "time/tzdata"

	_api "github.com/BANKA2017/tbsign_go/api"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	_type "github.com/BANKA2017/tbsign_go/types"
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
var autoInstall bool
var _adminName string
var _adminEmail string
var _adminPassword string

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

	// setup
	flag.BoolVar(&setup, "setup", false, "Init the system [force]")
	flag.BoolVar(&autoInstall, "auto_install", false, "Auto install the system when tables are not exist")
	flag.StringVar(&_adminName, "admin_name", "", "Name of admin")
	flag.StringVar(&_adminEmail, "admin_email", "", "Email of admin")
	flag.StringVar(&_adminPassword, "admin_password", "", "Password of admin")

	// others
	flag.BoolVar(&testMode, "test", false, "Not send any requests to tieba servers")

	flag.Parse()

	if setup {
		log.Println("WARNING: 覆盖安装已启用，会覆盖现有数据，请做好备份")
	}

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

	if !autoInstall && os.Getenv("tc_auto_install") != "" {
		autoInstall = true
	}
	if _adminName == "" && os.Getenv("tc_admin_name") != "" {
		_adminName = os.Getenv("tc_admin_name")
	}
	if _adminEmail == "" && os.Getenv("tc_admin_email") != "" {
		_adminEmail = os.Getenv("tc_admin_email")
	}
	if _adminPassword == "" && os.Getenv("tc_admin_password") != "" {
		_adminPassword = os.Getenv("tc_admin_password")
	}

	if setup && autoInstall {
		log.Fatal("ERROR: 不允许自动化覆盖安装!!!")
	} else if autoInstall && _adminName != "" && _adminEmail != "" && _adminPassword != "" {
		log.Println("WARNING: 已启用自动安装")
	} else if autoInstall {
		log.Fatal("ERROR: 管理员信息不完整，无法安装")
	}

	// connect to db
	dbMode := "mysql"
	logLevel := logger.Error
	if testMode {
		logLevel = logger.Info
	}

	dbExists := true

	if dbPath != "" {
		// sqlite
		dbMode = "sqlite"
		if _, err := os.Stat(dbPath); err != nil && os.IsNotExist(err) {
			dbExists = false
			setup = true
		}
		_function.GormDB.R, _function.GormDB.W, err = _function.ConnectToSQLite(dbPath, logLevel, "tbsign")
		if err != nil {
			log.Fatal("db:", err)
		}

		// setup
		if setup {
			_function.SetupSystem(dbMode, dbPath, "", "", "", "", logLevel, dbExists, _tc_mysql, _tc_sqlite, _tc_init_system, autoInstall, _adminName, _adminEmail, _adminPassword)
		}
	} else {
		// mysql
		if dbUsername == "" || dbPassword == "" {
			log.Fatal("global: Empty username or password")
		}
		// precheck table
		_function.GormDB.R, _function.GormDB.W, err = _function.ConnectToMySQL(dbUsername, dbPassword, dbEndpoint, "", logLevel, "db")

		if err != nil {
			log.Fatal("db:", err)
		}

		var count struct {
			Count int64
		}

		_function.GormDB.R.Raw("SELECT (COUNT(*) > 0) AS count FROM information_schema.tables WHERE table_schema = ?;", dbName).Scan(&count)
		dbExists = count.Count > 0
		if !dbExists {
			log.Println("db:", dbName, "is not exists")
			setup = true
		}

		// setup
		if setup {
			_function.SetupSystem(dbMode, "", dbUsername, dbPassword, dbEndpoint, dbName, logLevel, dbExists, _tc_mysql, _tc_sqlite, _tc_init_system, autoInstall, _adminName, _adminEmail, _adminPassword)
		} else {
			_function.GormDB.R, _function.GormDB.W, err = _function.ConnectToMySQL(dbUsername, dbPassword, dbEndpoint, dbName, logLevel, "db")
			if err != nil {
				log.Fatal("db:", err)
			}
		}
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
