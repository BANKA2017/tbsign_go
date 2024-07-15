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
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"gorm.io/gorm/logger"
)

// install
var setup bool
var autoInstall bool

var _adminName string
var _adminEmail string
var _adminPassword string

var err error

func main() {
	// sqlite
	flag.StringVar(&share.DBPath, "db_path", "", "Database path")

	// mysql
	flag.StringVar(&share.DBUsername, "username", "", "Username")
	flag.StringVar(&share.DBPassword, "pwd", "", "Password")
	flag.StringVar(&share.DBEndpoint, "endpoint", "127.0.0.1:3306", "endpoint")
	flag.StringVar(&share.DBName, "db", "tbsign", "Database name")

	//proxy
	flag.BoolVar(&_function.IgnoreProxy, "no_proxy", false, "Ignore the http proxy config from environment vars")

	// api
	flag.BoolVar(&share.EnableApi, "api", false, "active backend endpoints")
	flag.BoolVar(&share.EnableFrontend, "fe", false, "active frontend endpoints")
	flag.StringVar(&share.Address, "address", ":1323", "address :1323")

	// setup
	flag.BoolVar(&setup, "setup", false, "Init the system [force]")
	flag.BoolVar(&autoInstall, "auto_install", false, "Auto install the system when tables are not exist")
	flag.StringVar(&_adminName, "admin_name", "", "Name of admin")
	flag.StringVar(&_adminEmail, "admin_email", "", "Email of admin")
	flag.StringVar(&_adminPassword, "admin_password", "", "Password of admin")

	// others
	flag.BoolVar(&share.TestMode, "test", false, "Not send any requests to tieba servers")

	flag.Parse()

	if setup {
		log.Println("WARNING: 覆盖安装已启用，会覆盖现有数据，请做好备份")
	}

	// from env
	if share.DBUsername == "" {
		share.DBUsername = os.Getenv("tc_username")
	}
	if share.DBPassword == "" {
		share.DBPassword = os.Getenv("tc_pwd")
	}
	if share.DBEndpoint == "" && os.Getenv("tc_endpoint") != "" {
		share.DBEndpoint = os.Getenv("tc_endpoint")
	}
	if share.DBName == "" && os.Getenv("tc_db") != "" {
		share.DBName = os.Getenv("tc_db")
	}
	if share.DBPath == "" && os.Getenv("tc_db_path") != "" {
		share.DBPath = os.Getenv("tc_db_path")
	}
	if !share.TestMode && os.Getenv("tc_test") != "" {
		share.TestMode = os.Getenv("tc_test") == "true"
	}
	if !share.EnableApi && os.Getenv("tc_api") != "" {
		share.EnableApi = os.Getenv("tc_api") == "true"
	}
	if !share.EnableFrontend && os.Getenv("tc_fe") != "" {
		share.EnableFrontend = os.Getenv("tc_fe") == "true"
	}

	if !share.EnableApi && share.EnableFrontend {
		log.Fatal("ERROR: 不允许关闭仅启用前端!!!")
	}

	if share.Address == ":1323" && os.Getenv("tc_address") != "" {
		share.Address = os.Getenv("tc_address")
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
	if share.TestMode {
		logLevel = logger.Info
	}

	dbExists := true

	if share.DBPath != "" {
		// sqlite
		dbMode = "sqlite"
		if _, err := os.Stat(share.DBPath); err != nil && os.IsNotExist(err) {
			dbExists = false
			setup = true
		}
		_function.GormDB.R, _function.GormDB.W, err = _function.ConnectToSQLite(share.DBPath, logLevel, "tbsign")
		if err != nil {
			log.Fatal("db:", err)
		}

		// setup
		if setup {
			_function.SetupSystem(dbMode, share.DBPath, "", "", "", "", logLevel, dbExists, autoInstall, _adminName, _adminEmail, _adminPassword)
		}
	} else {
		// mysql
		if share.DBUsername == "" || share.DBPassword == "" {
			log.Fatal("global: Empty username or password")
		}
		// precheck table
		_function.GormDB.R, _function.GormDB.W, err = _function.ConnectToMySQL(share.DBUsername, share.DBPassword, share.DBEndpoint, "", logLevel, "db")

		if err != nil {
			log.Fatal("db:", err)
		}

		var count struct {
			Count int64
		}

		_function.GormDB.R.Raw("SELECT (COUNT(*) > 0) AS count FROM information_schema.tables WHERE table_schema = ?;", share.DBName).Scan(&count)
		dbExists = count.Count > 0
		if !dbExists {
			log.Println("db:", share.DBName, "is not exists")
			setup = true
		}

		// setup
		if setup {
			_function.SetupSystem(dbMode, "", share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, logLevel, dbExists, autoInstall, _adminName, _adminEmail, _adminPassword)
		} else {
			_function.GormDB.R, _function.GormDB.W, err = _function.ConnectToMySQL(share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, logLevel, "db")
			if err != nil {
				log.Fatal("db:", err)
			}
		}
	}

	// init
	_function.GetOptionsAndPluginList()

	if share.EnableApi {
		go _api.Api(share.Address, "dbmode", dbMode, "testmode", share.TestMode)
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
			if share.TestMode {
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

			if p, ok := _function.PluginList["ver4_ref"]; ok && p.Status {
				go _plugin.RefreshTiebaListAction()
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
			_function.GetOptionsAndPluginList()

			// clean cookie/fid cache
			_function.CookieList = make(map[int32]_type.TypeCookie)
			_function.FidList = make(map[string]int64)
		}
	}
}
