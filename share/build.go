package share

import (
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
)

func init() {
	if MockReleasedVersion && BuildPublishType == "source" {
		BuiltAt = time.Now().Format(time.RFC3339)
		BuildRuntime = runtime.GOOS + "/" + runtime.GOARCH
		BuildGitCommitHash = "be27116224bfea617de0c7f1f00d6dff0afd5d62"
		BuildEmbeddedFrontendGitCommitHash = "0dc1f57e3495e21803152460588e1fd010dcf5c0"
		BuildPublishType = "binary"
	}

	t, err := time.Parse(time.RFC3339, BuiltAt)
	if err != nil {
		BuildAtTime = time.Now()
	} else {
		BuildAtTime = t
	}

	// docker
	if _, err := os.Stat("/.dockerenv"); err == nil && os.Getenv("tc_docker_mode") == "true" {
		BuildPublishType = "docker"
	}

	// fe-hash
	indexFile, err := assets.EmbeddedFrontend.ReadFile("dist/index.html")

	if err != nil {
		return
	}

	re := regexp.MustCompile(`NUXT_COMMIT_HASH:"([0-9a-f]+)"`)
	m := re.FindStringSubmatch(string(indexFile))

	if len(m) > 1 {
		BuildEmbeddedFrontendGitCommitHash2 = m[1]
	}

	if BuildEmbeddedFrontendGitCommitHash2 == "" || BuildEmbeddedFrontendGitCommitHash2 != BuildEmbeddedFrontendGitCommitHash {
		slog.Warn("Frontend invalid or not built by current source code (share.build)", "embedded_fe_commit_hash", BuildEmbeddedFrontendGitCommitHash, "actual_fe_commit_hash", BuildEmbeddedFrontendGitCommitHash2)
	}
}

var BuiltAt = "Now"
var BuildRuntime = "Dev"
var BuildGitCommitHash = "N/A"
var BuildEmbeddedFrontendGitCommitHash = "N/A"
var BuildEmbeddedFrontendGitCommitHash2 = ""
var BuildPublishType = "source"

var BuildAtTime time.Time

var ReleaseFilesPath = "https://github.com/BANKA2017/tbsign_go/releases/download"

// ReleaseApiBase https://api.github.com/repos/{owner}/{repo}/releases/tags/{tag_id} // <-
// see also: https://docs.github.com/zh/rest/releases/releases?apiVersion=2022-11-28#get-a-release
var ReleaseApiBase = "https://api.github.com/repos/BANKA2017/tbsign_go/releases/tags/"
