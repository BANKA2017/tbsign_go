package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var dbUsername string
var dbPassword string
var dbEndpoint string
var dbName string

var dbPath string

func main() {
	// sqlite
	flag.StringVar(&dbPath, "db_path", "tbsign.db", "Database path")

	// mysql
	flag.StringVar(&dbUsername, "username", "", "Username")
	flag.StringVar(&dbPassword, "pwd", "", "Password")
	flag.StringVar(&dbEndpoint, "endpoint", "127.0.0.1:3306", "endpoint")
	flag.StringVar(&dbName, "db", "tbsign", "Database name")

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

	// connect to db

	if _, err := os.Stat(dbPath); err == nil {
		// sqlite
		_function.GormDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

		if err != nil {
			log.Fatal("db:", err)
		}
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

	// Interval
	oneMinuteInterval := time.NewTicker(time.Minute)
	defer oneMinuteInterval.Stop()
	fourHoursInterval := time.NewTicker(time.Hour * 4)
	defer fourHoursInterval.Stop()

	for {
		select {
		case <-oneMinuteInterval.C:
			_function.GetOptionsAndPluginList()

			_function.UpdateNow()
			_plugin.DoSignAction()
			_plugin.DoReSignAction()

			// plugins
			if _function.PluginList["ver4_rank"] {
				go _plugin.DoForumSupportAction()
			}

			if _function.PluginList["ver4_ban"] {
				go _plugin.LoopBanAction()
			}
		case <-fourHoursInterval.C:
			_function.GetOptionsAndPluginList()
			if _function.PluginList["ver4_ref"] {
				go _plugin.RefreshTiebaListAction()
			}
		}
	}
}
