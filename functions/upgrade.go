package _function

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/kdnetwork/code-snippet/go/utils"
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

var passList = map[string][]string{
	"darwin":  {"amd64", "arm64"},
	"linux":   {"amd64", "arm64"},
	"windows": {"amd64"},
}

func IsOfficialSupport() bool {
	list, ok := passList[runtime.GOOS]
	return ok && slices.Contains(list, runtime.GOARCH)
}

func IsBinaryType() bool {
	return strings.ToLower(share.BuildPublishType) == "binary"
}

var ReleaseFilesPath = share.ReleaseFilesPath

type chunk struct {
	start int
	end   int
	index int
	data  []byte
	err   error
}

var defaultHeaders = map[string]string{
	"User-Agent": "tbsign_go/upgrader",
}

// by chatgpt

func formatSpeedMB(bytes int64, duration time.Duration) string {
	if duration <= 0 {
		return "∞ MB/s"
	}
	mbPerSec := float64(bytes) / 1024 / 1024 / duration.Seconds()
	return fmt.Sprintf("%.2f MB/s", mbPerSec)
}
func upgradeDownloader(_url string, _client *http.Client) ([]byte, error) {
	// get response headers
	req, err := http.NewRequest("HEAD", _url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", defaultHeaders["User-Agent"])

	resp, err := _client.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	lengthStr := resp.Header.Get("Content-Length")
	if lengthStr == "" {
		return Fetch(_url, http.MethodGet, nil, defaultHeaders, _client)
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil || length <= 0 {
		return Fetch(_url, http.MethodGet, nil, defaultHeaders, _client)
	}

	// Range?
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		return Fetch(_url, http.MethodGet, nil, defaultHeaders, _client)
	}

	workers := utils.Clamp(runtime.NumCPU(), 4, 8)

	chunkSize := length / workers

	chunks := make([]*chunk, workers)

	var wg sync.WaitGroup
	wg.Add(workers)

	// concurrent
	for i := range workers {
		start := i * chunkSize
		end := start + chunkSize - 1

		if i == workers-1 {
			end = length - 1
		}

		chunks[i] = &chunk{
			start: start,
			end:   end,
			index: i,
		}

		go func(c *chunk) {
			defer wg.Done()

			data, err := Fetch(_url, http.MethodGet, nil, map[string]string{
				"User-Agent": defaultHeaders["User-Agent"],
				"Range":      fmt.Sprintf("bytes=%d-%d", c.start, c.end),
			}, _client)

			if err != nil {
				c.err = err
				return
			}

			c.data = data
		}(chunks[i])
	}

	wg.Wait()

	// check
	for _, c := range chunks {
		if c.err != nil {
			return nil, c.err
		}
	}

	// concat chunks
	var buf bytes.Buffer
	for _, c := range chunks {
		buf.Write(c.data)
	}

	return buf.Bytes(), nil
}

// version = "20240707.c7990c7.6a6db54"
func Upgrade(version string) error {
	_os := runtime.GOOS
	_arch := runtime.GOARCH

	if share.BuiltAt == "Now" {
		return errors.New("❌ 不支持的版本 (开发版)，请参考 build.sh 编译运行")
	} else if !IsBinaryType() {
		return fmt.Errorf("❌ 不支持直接下载更新的版本(%s)", share.BuildPublishType)
	} else if !IsOfficialSupport() {
		return fmt.Errorf("❌ 不支持的版本(%s/%s)，请下载源码后参考 build.sh 编译运行", runtime.GOOS, runtime.GOARCH)
	} else if _, err := url.Parse(ReleaseFilesPath); err != nil {
		return errors.New("❌ 更新地址无效")
	}

	// pre check version
	if !regexp.MustCompile(`^\d{8}\.[0-9a-f]{7}\.[0-9a-f]{7}$`).MatchString(version) {
		return errors.New("❌ 版本号格式不正确")
	}

	//get releases "https://api.github.com/repos/banka2017/tbsign_go/releases?per_page=5"

	releasePath, _ := url.JoinPath(ReleaseFilesPath, "tbsign_go."+version)

	binPath, _ := url.JoinPath(releasePath, "tbsign_go."+version+"."+_os+"-"+_arch)

	if _os == "windows" {
		binPath += ".exe"
	}

	sha256Path, _ := url.JoinPath(releasePath, "tbsign_go."+version+"."+_os+"-"+_arch+".sha256")
	if _os == "windows" {
		sha256Path = strings.ReplaceAll(sha256Path, ".sha256", ".exe.sha256")
	}

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	slog.Info("replace file", "path", execPath)

	// Path to the new version temporary file
	tmpFile := filepath.Join(os.TempDir(), "__tmp__tbsign-binary.tmp")

	// get binary
	since := time.Now()
	binary, err := upgradeDownloader(binPath, DefaultClient)
	if err != nil {
		return err
	}

	duration := time.Since(since)
	slog.Info("download status", "duration", int(duration.Seconds()), "avg-speed", formatSpeedMB(int64(len(binary)), duration))

	sha256Str, err := Fetch(sha256Path, http.MethodGet, nil, defaultHeaders, DefaultClient)
	if err != nil {
		return err
	}

	return verifyAndSave(binary, execPath, tmpFile, string(sha256Str), len(binary))
}

// https://api.github.com/repos/${{repo}}/releases/tags/${{tag}}
type releaseInfoStruct struct {
	ID          int       `json:"id"`
	TagName     string    `json:"tag_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		URL                string    `json:"url"`
		ID                 int       `json:"id"`
		Name               string    `json:"name"` // <-
		State              string    `json:"state"`
		Size               int       `json:"size"`   // <-
		Digest             string    `json:"digest"` // <-
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"` // <-
	} `json:"assets"`
	Body string `json:"body"`
}

func Upgrade2(tagName string) error {
	_os := runtime.GOOS
	_arch := runtime.GOARCH

	if share.BuiltAt == "Now" {
		return errors.New("❌ 不支持的版本 (开发版)，请参考 build.sh 编译运行")
	} else if !IsBinaryType() {
		return fmt.Errorf("❌ 不支持直接下载更新的版本(%s)", share.BuildPublishType)
	} else if !IsOfficialSupport() {
		return fmt.Errorf("❌ 不支持的版本(%s/%s)，请下载源码后参考 build.sh 编译运行", runtime.GOOS, runtime.GOARCH)
	} else if _, err := url.Parse(share.ReleaseApiBase + "/releases/tags/"); err != nil {
		return errors.New("❌ 更新地址无效")
	}

	// pre check tagName
	if !regexp.MustCompile(`^tbsign_go\.\d{8}\.[0-9a-f]{7}\.[0-9a-f]{7}$`).MatchString(tagName) {
		return errors.New("❌ 版本号格式不正确")
	}

	// get info
	binInfo, err := Fetch(share.ReleaseApiBase+"/releases/tags/"+tagName, http.MethodGet, nil, map[string]string{
		"User-Agent": "tbsign_go/upgrader2",
	}, DefaultClient)

	var info releaseInfoStruct

	err = JsonDecode(binInfo, &info)
	if err != nil {
		return err
	}

	if info.TagName != tagName {
		return errors.New("❌ 版本号不匹配")
	} else if len(info.Assets) == 0 {
		return errors.New("❌ 无可用二进制文件")
	} else if share.BuildAtTime.After(info.PublishedAt) && GetOption("go_allow_downgrade") != "1" {
		return errors.New("❌ 版本已过期")
	}

	// find binary
	binName := tagName + "." + _os + "-" + _arch

	if _os == "windows" {
		binName += ".exe"
	}

	var binPath string
	var fileSha256 string
	var fileSize int

	for _, asset := range info.Assets {
		if asset.Name == binName {
			binPath = asset.BrowserDownloadURL
			fileSha256 = strings.TrimSpace(strings.ReplaceAll(asset.Digest, "sha256:", ""))
			fileSize = asset.Size
			break
		}
	}
	if binPath == "" {
		return errors.New("❌ 未找到匹配的二进制文件")
	}

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	slog.Info("replace file", "path", execPath)

	// Path to the new tagName temporary file
	tmpFile := filepath.Join(os.TempDir(), "__tmp__tbsign-binary.tmp")

	// get binary
	since := time.Now()
	binary, err := upgradeDownloader(binPath, DefaultClient)
	if err != nil {
		return err
	}

	duration := time.Since(since)
	slog.Info("download status", "duration", int(duration.Seconds()), "avg-speed", formatSpeedMB(int64(len(binary)), duration))

	return verifyAndSave(binary, execPath, tmpFile, fileSha256, fileSize)
}

func verifyAndSave(binary []byte, execPath, tmpPath, fileSha256 string, fileSize int) error {
	// diff sha256
	hasher := sha256.Sum256(binary)
	hashString := strings.ToLower(strings.TrimSpace(hex.EncodeToString(hasher[:])))

	slog.Info("File info", "size", len(binary), "calculated-sha256", hashString, "expected-sha256", fileSha256)

	if hashString == strings.ToLower(strings.TrimSpace(fileSha256)) && len(binary) == fileSize {
		out, err := os.Create(tmpPath)
		if err != nil {
			return err
		}
		defer out.Close()

		file, err := os.Open(tmpPath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(out, bytes.NewReader(binary))
		if err != nil {
			return err
		}

		s, err := file.Stat()
		if err != nil {
			return err
		}
		if s.Size() != int64(fileSize) {
			return errors.New("更新失败！文件大小不匹配")
		}

		if runtime.GOOS != "windows" {
			os.Chmod(tmpPath, 0755)
			err = os.Rename(tmpPath, execPath)
			if err != nil {
				return err
			}
		} else {
			win_upgrade_script_template, _ := assets.EmbeddedUpgradeFiles.ReadFile("upgrade/win_upgrade_script_template.ps1")

			psScript := fmt.Sprintf(string(win_upgrade_script_template), execPath, tmpPath)

			psFile := filepath.Join(os.TempDir(), "tc_win_upgrade_script.ps1")
			if err := os.WriteFile(psFile, []byte(psScript), 0644); err != nil {
				return err
			}

			cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", psFile)
			if err := cmd.Start(); err != nil {
				return err
			}
		}

		slog.Info("更新完成")
	} else {
		return errors.New("更新失败！sha256 记录或文件大小不正确")
	}
	return nil
}
