package share

import "time"

func init() {
	t, err := time.Parse(time.RFC3339, BuiltAt)
	if err != nil {
		BuildAtTime = time.Now()
	} else {
		BuildAtTime = t
	}
}

var BuiltAt = "Now"
var BuildRuntime = "Dev"
var BuildGitCommitHash = "N/A"
var BuildEmbeddedFrontendGitCommitHash = "N/A"
var BuildPublishType = "source"

var BuildAtTime time.Time

var ReleaseFilesPath = "https://github.com/BANKA2017/tbsign_go/releases/download"

// ReleaseApiBase https://api.github.com/repos/{owner}/{repo}/releases/tags/{tag_id} // <-
// see also: https://docs.github.com/zh/rest/releases/releases?apiVersion=2022-11-28#get-a-release
var ReleaseApiBase = "https://api.github.com/repos/BANKA2017/tbsign_go/releases/tags/"
