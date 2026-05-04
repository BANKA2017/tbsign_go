package share

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
)

func init() {
	if MockReleasedVersion && BuildPublishType == "source" {
		BuildPublishType = "binary"
	}

	// vcs info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		kv := make(map[string]string)
		for _, s := range buildInfo.Settings {
			kv[s.Key] = s.Value
		}

		BuildGitCommitHash = kv["vcs.revision"]
		BuildDirty = kv["vcs.modified"] == "true"
		BuildGitCommitTime = kv["vcs.time"]

		if BuiltAt == "" {
			BuiltAt = BuildGitCommitTime
		}
	}

	// build at
	t, err := time.Parse(time.RFC3339, BuiltAt)
	if err != nil {
		BuildAtTime = time.Now()
		BuiltAt = BuildAtTime.Format(time.RFC3339)
	} else {
		BuildAtTime = t
	}

	// build runtime
	if BuildRuntime == "" {
		BuildRuntime = runtime.GOOS + "/" + runtime.GOARCH
	}

	// docker
	if _, err := os.Stat("/.dockerenv"); err == nil && os.Getenv("tc_docker_mode") == "true" {
		BuildPublishType = "docker"
	}

	// fe-hash
	indexFile, err := assets.EmbeddedFrontend.ReadFile("dist/index.html")

	if err == nil {
		re := regexp.MustCompile(`NUXT_COMMIT_HASH:"([0-9a-f]+)"`)
		m := re.FindStringSubmatch(string(indexFile))

		if len(m) > 1 {
			BuildEmbeddedFrontendGitCommitHash2 = m[1]
		}

		// if BuildEmbeddedFrontendGitCommitHash2 == "" || BuildEmbeddedFrontendGitCommitHash2 != BuildEmbeddedFrontendGitCommitHash {
		// 	slog.Warn("Frontend invalid or not built by current source code (share.build)", "embedded_fe_commit_hash", BuildEmbeddedFrontendGitCommitHash, "actual_fe_commit_hash", BuildEmbeddedFrontendGitCommitHash2)
		// }
		//
		// if BuildEmbeddedFrontendGitCommitHash == "" {
		BuildEmbeddedFrontendGitCommitHash = BuildEmbeddedFrontendGitCommitHash2
		// }
	}

	if len(BuildGitCommitHash) >= 7 && len(BuildEmbeddedFrontendGitCommitHash) >= 7 {
		DynamicVersion = fmt.Sprintf("%s.%s.%s", BuildAtTime.Format("20060102"), BuildGitCommitHash[0:7], BuildEmbeddedFrontendGitCommitHash[0:7])
	}
}

var BuiltAt = ""
var BuildRuntime = ""
var BuildGitCommitHash = ""
var BuildGitCommitTime = ""
var BuildEmbeddedFrontendGitCommitHash = ""
var BuildEmbeddedFrontendGitCommitHash2 = ""
var BuildPublishType = "source"
var BuildDirty = true

var BuildAtTime time.Time

var ReleaseFilesPath = "https://github.com/BANKA2017/tbsign_go/releases/download"

// ReleaseApiBase https://api.github.com/repos/{owner}/{repo}/releases/tags/{tag_id} // <-
// see also: https://docs.github.com/zh/rest/releases/releases?apiVersion=2022-11-28#get-a-release
var ReleaseApiBase = "https://api.github.com/repos/BANKA2017/tbsign_go/releases/tags/"
var ReleaseApiList = "https://api.github.com/repos/BANKA2017/tbsign_go/releases"

var DynamicVersion = "dev"
