package _function

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/BANKA2017/tbsign_go/share"
	"golang.org/x/exp/slices"
)

type GithubReleasesListResponseItem struct {
	HTMLURL         string `json:"html_url,omitempty"`
	ID              int    `json:"id,omitempty"`
	TagName         string `json:"tag_name,omitempty"`
	TargetCommitish string `json:"target_commitish,omitempty"`
	Name            string `json:"name,omitempty"`
	Draft           bool   `json:"draft,omitempty"`
	Prerelease      bool   `json:"prerelease,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
	Assets          []struct {
		URL                string `json:"url,omitempty"`
		Name               string `json:"name,omitempty"`
		State              string `json:"state,omitempty"`
		Size               int    `json:"size,omitempty"`
		CreatedAt          string `json:"created_at,omitempty"`
		UpdatedAt          string `json:"updated_at,omitempty"`
		BrowserDownloadURL string `json:"browser_download_url,omitempty"`
	} `json:"assets,omitempty"`
	Body string `json:"body,omitempty"`
}

var passList map[string][]string = map[string][]string{
	"darwin":  {"amd64", "arm64"},
	"linux":   {"amd64", "arm64"},
	"windows": {"amd64"},
}

func IsOfficialSupport() bool {
	if list, ok := passList[runtime.GOOS]; ok {
		if slices.Contains(list, runtime.GOARCH) {
			return true
		}
	}
	return false
}

//TODO upgrade

// version = "20240707.c7990c7.6a6db54"
func Upgrade(version string) error {
	_os := runtime.GOOS
	_arch := runtime.GOARCH

	if !IsOfficialSupport() {
		return fmt.Errorf("❌ 不支持的版本(%s/%s)，请下载源码后参考 build.sh 编译运行", runtime.GOOS, runtime.GOARCH)
	} else if share.BuiltAt == "Now" {
		return fmt.Errorf("❌ 不支持的版本 (开发版)，请参考 build.sh 编译运行")
	}

	// pre check version
	if !regexp.MustCompile(`^\d{8}\.[0-9a-f]{7}\.[0-9a-f]{7}$`).MatchString(version) {
		return fmt.Errorf("❌ 版本号格式不正确")
	}

	//get releases "https://api.github.com/repos/banka2017/tbsign_go/releases?per_page=5"

	binPath := "https://github.com/BANKA2017/tbsign_go/releases/download/tbsign_go." + version + "/tbsign_go." + version + "." + _os + "-" + _arch
	sha256Path := "https://github.com/BANKA2017/tbsign_go/releases/download/tbsign_go." + version + "/tbsign_go." + version + "." + _os + "-" + _arch + ".sha256"

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	log.Println("更新将会替换掉当前运行文件:", execPath)
	execDir := filepath.Dir(execPath)

	// Path to the new version temporary file
	tmpFile := filepath.Join(execDir, "__tmp__tbsign.tmp")

	// get binary
	binary, err := Fetch(binPath, "GET", nil, map[string]string{
		"User-Agent": BrowserUserAgent,
	}, DefaultCient)
	if err != nil {
		return err
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	file, err := os.Open(tmpFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(out, bytes.NewReader(binary))
	if err != nil {
		return err
	}

	// diff sha256
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	s, _ := file.Stat()

	sha256Str, err := Fetch(sha256Path, "GET", nil, map[string]string{
		"User-Agent": BrowserUserAgent,
	}, DefaultCient)
	if err != nil {
		return err
	}

	log.Println(s.Size(), strings.TrimSpace(hashString), strings.TrimSpace(string(sha256Str)))

	if strings.TrimSpace(hashString) == strings.TrimSpace(string(sha256Str)) {
		os.Chmod(tmpFile, 0755)
		err = os.Rename(tmpFile, execPath)
		if err != nil {
			return err
		}

		fmt.Println("更新完成！")
	} else {
		os.Remove(tmpFile)
		return fmt.Errorf("更新失败！sha256 记录不正确")
	}
	return nil
}
