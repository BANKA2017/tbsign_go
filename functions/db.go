package _function

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var GormDB = new(GormDBPool)

type GormDBPool struct {
	R *gorm.DB
	W *gorm.DB
}

func ConnectToSQLite(path string, logLevel logger.LogLevel, servicePrefix string) (*gorm.DB, *gorm.DB, error) {
	var writeDBHandle = new(gorm.DB)
	var readDBHandle = new(gorm.DB)
	var err error
	if _, err = os.Stat(path); err != nil {
		log.Println("db:", path, "is not exists")
	}
	// sqlite
	// write

	writeDBHandle, err = gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Println(servicePrefix+" w:", err)
	}
	connw, err := writeDBHandle.DB()
	connw.SetMaxOpenConns(1)

	if err != nil {
		log.Println(servicePrefix+" w:", err)
	}

	//read
	readDBHandle, err = gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Println(servicePrefix+" r:", err)
	}
	// connr, err := readDBHandle.DB()
	// connr.SetMaxOpenConns(max(4, runtime.NumCPU()))

	if err != nil {
		log.Println(servicePrefix+" r:", err)
	}
	log.Println(servicePrefix + ": sqlite connected!")

	writeDBHandle.Exec("PRAGMA journal_mode = WAL;PRAGMA busy_timeout = 5000;PRAGMA synchronous = NORMAL;PRAGMA cache_size = 100000;PRAGMA foreign_keys = true;PRAGMA temp_store = memory;")
	return readDBHandle, writeDBHandle, err
}

func ConnectToMySQL(username string, password string, endpoint string, dbname string, logLevel logger.LogLevel, servicePrefix string) (*gorm.DB, *gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, endpoint, dbname)

	sqlDB, _ := sql.Open("mysql", dsn)
	//defer sqlDB.Close()
	dbHandle, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{Logger: logger.Default.LogMode(logLevel)})

	log.Println(servicePrefix, ": mysql connected!")
	return dbHandle, dbHandle, err
}
