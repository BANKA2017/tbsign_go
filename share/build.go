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
