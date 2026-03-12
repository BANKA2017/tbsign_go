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
	"strings"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
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

	binPath := ReleaseFilesPath + "/tbsign_go." + version + "/tbsign_go." + version + "." + _os + "-" + _arch

	if _os == "windows" {
		binPath += ".exe"
	}

	sha256Path := ReleaseFilesPath + "/tbsign_go." + version + "/tbsign_go." + version + "." + _os + "-" + _arch + ".sha256"
	if _os == "windows" {
		sha256Path = strings.ReplaceAll(sha256Path, ".sha256", ".exe.sha256")
	}

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	slog.Info("replace file", "path", execPath)
	execDir := filepath.Dir(execPath)

	// Path to the new version temporary file
	tmpFile := filepath.Join(execDir, "__tmp__tbsign.tmp")

	// get binary
	binary, err := Fetch(binPath, http.MethodGet, nil, map[string]string{
		"User-Agent": "tbsign_go/upgrader",
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

	sha256Str, err := Fetch(sha256Path, http.MethodGet, nil, map[string]string{
		"User-Agent": "tbsign_go/upgrader",
	}, DefaultCient)
	if err != nil {
		return err
	}

	slog.Info("File info", "size", s.Size(), "calculated-sha256", strings.TrimSpace(hashString), "expected-sha256", strings.TrimSpace(string(sha256Str)))

	if strings.TrimSpace(hashString) == strings.TrimSpace(string(sha256Str)) {
		if _os != "windows" {
			os.Chmod(tmpFile, 0755)
			err = os.Rename(tmpFile, execPath)
			if err != nil {
				return err
			}
		} else {
			win_upgrade_script_template, _ := assets.EmbeddedUpgradeFiles.ReadFile("upgrade/win_upgrade_script_template.ps1")

			psScript := fmt.Sprintf(string(win_upgrade_script_template), execPath, tmpFile)

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
		os.Remove(tmpFile)
		return errors.New("更新失败！sha256 记录不正确")
	}
	return nil
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
	} else if _, err := url.Parse(share.ReleaseApiBase); err != nil {
		return errors.New("❌ 更新地址无效")
	}

	// pre check tagName
	if !regexp.MustCompile(`^tbsign_go\.\d{8}\.[0-9a-f]{7}\.[0-9a-f]{7}$`).MatchString(tagName) {
		return errors.New("❌ 版本号格式不正确")
	}

	// get info
	binInfo, err := Fetch(share.ReleaseApiBase+tagName, http.MethodGet, nil, map[string]string{
		"User-Agent": "tbsign_go/upgrader2",
	}, DefaultCient)

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
	execDir := filepath.Dir(execPath)

	// Path to the new tagName temporary file
	tmpFile := filepath.Join(execDir, "__tmp__tbsign.tmp")

	// get binary
	binary, err := Fetch(binPath, http.MethodGet, nil, map[string]string{
		"User-Agent": "tbsign_go/upgrader2",
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

	slog.Info("File info", "size", s.Size(), "calculated-sha256", strings.TrimSpace(hashString), "expected-sha256", fileSha256)

	if strings.TrimSpace(hashString) == fileSha256 && s.Size() == int64(fileSize) {
		if _os != "windows" {
			os.Chmod(tmpFile, 0755)
			err = os.Rename(tmpFile, execPath)
			if err != nil {
				return err
			}
		} else {
			win_upgrade_script_template, _ := assets.EmbeddedUpgradeFiles.ReadFile("upgrade/win_upgrade_script_template.ps1")

			psScript := fmt.Sprintf(string(win_upgrade_script_template), execPath, tmpFile)

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
		os.Remove(tmpFile)
		return errors.New("更新失败！sha256 记录或文件大小不正确")
	}
	return nil
}
