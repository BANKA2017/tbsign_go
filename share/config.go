package share

import "time"

var DBUsername string
var DBPassword string
var DBEndpoint string
var DBName string
var DBTLSOption string
var DBVersion string

var DBPath string
var DBMode string

var TestMode bool
var EnableApi bool
var EnableFrontend bool
var EnableBackup bool

var Address string

var StartTime = time.Now()

var IsPureGO bool
var IsEncrypt bool

// --experimental-*
var DataEncryptKeyStr string
var DataEncryptKeyByte []byte
var DisableEmail bool
var DNSAddress string
