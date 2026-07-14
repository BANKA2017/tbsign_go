package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	_api "github.com/BANKA2017/tbsign_go/api"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/kdnetwork/code-snippet/go/log"
	"github.com/kdnetwork/code-snippet/go/utils"
)

var err error

func init() {
	_function.InitDefaultLogger()

	if !utils.GetBoolEnv("tc_hide_version_info") {
		fmt.Println("TbSign➡️\n--- info ---")
		fmt.Println("build_at:", share.BuiltAt)
		fmt.Println("build_runtime:", runtime.Version()+" "+share.BuildRuntime)
		fmt.Println("commit_hash:", share.BuildGitCommitHash)
		fmt.Println("frontend_hash:", share.BuildEmbeddedFrontendGitCommitHash)
		fmt.Println("release_type:", share.BuildPublishType)
		fmt.Println("version:", share.DynamicVersion)
		fmt.Println("dirty_build:", share.BuildDirty)
		fmt.Print("------------\n\n")
	}
}

func main() {
	InitEnv()

	/// client
	/// DO NOT EXEC _function.InitClient BEFORE READING FLAGS AND ENV!!!!!
	_function.DefaultClient = _function.InitClient(30 * time.Minute)
	_function.TBClient = _function.InitClient(10 * time.Second)

	// connect to db
	InitDB()

	// init
	_function.InitOptions()
	share.IsPureGO = _function.GetOption("go_ver") == "1"

	_plugin.InitPluginList()

	InitEncrypt()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if share.EnableApi {
		go _api.Api(ctx, share.Network, share.Address)
	}

	// wait for the next minute
	if _function.GetOption("go_wait_for_next_minute_on_startup") != "0" {
		now := time.Now()
		nextMinute := now.Truncate(time.Minute).Add(time.Minute)
		slog.Info("等待到下一个整分钟以启动", "next_minute", nextMinute)
		time.Sleep(nextMinute.Sub(now))
	}

	// cron
	_function.Crontab, err = InitCrontab()
	if err != nil {
		log.Fatal("cron", "error", err)
	}

	// start the scheduler
	_function.Crontab.Start()

	defer _function.Crontab.Shutdown()

	<-ctx.Done()
}
