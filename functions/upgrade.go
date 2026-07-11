package _function

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"debug/buildinfo"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/goccy/go-yaml"
	ghup "github.com/kdnetwork/ghup/go-sdk"
	"golang.org/x/exp/slices"
)

func init() {
	pk, err := assets.EmbeddedCACert.ReadFile("ca/tc_p256_verify_public_v1.pem")
	if err != nil {
		return
	}

	block, _ := pem.Decode(pk)
	if block == nil {
		return
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return
	}

	VerifyPublicKey, _ = pubKeyInterface.(*ecdsa.PublicKey)

	UpgraderDefaultHeaders["User-Agent"] = "tbsign_go/upgrader2"
}

var VerifyPublicKey *ecdsa.PublicKey

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

var UpgraderDefaultHeaders = make(map[string]string)

// https://api.github.com/repos/${{repo}}/releases/tags/${{tag}}

func Upgrade2(tagName string) (*ghup.UpdateContent, error) {
	if share.BuiltAt == "Now" {
		return nil, errors.New("❌ 不支持的版本 (开发版)，请参考 build.sh 编译运行")
	} else if !IsBinaryType() {
		return nil, fmt.Errorf("❌ 不支持直接下载更新的版本(%s)", share.BuildPublishType)
	} else if !IsOfficialSupport() {
		return nil, fmt.Errorf("❌ 不支持的版本(%s/%s)，请下载源码后参考 build.sh 编译运行", runtime.GOOS, runtime.GOARCH)
	} else if _, err := url.Parse(share.ReleaseApiBase + "/releases/tags/"); err != nil {
		return nil, errors.New("❌ 更新地址无效")
	}

	// pre-check tagName
	if !regexp.MustCompile(`^tbsign_go\.\d{8}\.[0-9a-f]{7}\.[0-9a-f]{7}$`).MatchString(tagName) {
		return nil, errors.New("❌ 版本号格式不正确")
	}

	uCtx := ghup.NewUpgrader(share.ReleaseRepo)

	uCtx.APIPrefix = share.ReleaseApiBase
	uCtx.Client = DefaultClient
	uCtx.Headers = UpgraderDefaultHeaders

	if err := uCtx.Upgrade2(tagName); err != nil {
		return nil, err
	}

	if err := uCtx.Verify(); err != nil {
		return nil, err
	}

	return uCtx, nil
}

type releaseMetaData struct {
	BuildVersion string            `yaml:"build_version"`
	BuildAt      string            `yaml:"build_at"`
	CommitHash   string            `yaml:"commit_hash"`
	FrontendHash string            `yaml:"frontend_hash"`
	PublishType  string            `yaml:"publish_type"`
	Version      string            `yaml:"version"`
	Sha256       map[string]string `yaml:"sha256"`
	Signature    string            `yaml:"signature"`
}

func Upgrade3(binary []byte, metadata string) (*ghup.UpdateContent, error) {
	// public key
	if VerifyPublicKey == nil {
		return nil, errors.New("❌ 验证公钥无效")
	}

	// release metadata
	var release releaseMetaData

	if err := yaml.Unmarshal([]byte(metadata), &release); err != nil {
		return nil, errors.New("❌ 解析元数据失败")
	}

	splitMetadata := strings.Split(strings.TrimSpace(metadata), "\n")
	data := strings.Join(splitMetadata[0:len(splitMetadata)-1], "\n") + "\n"
	sig, err := base64.RawURLEncoding.DecodeString(release.Signature)
	if err != nil {
		return nil, errors.New("❌ 解析签名失败")
	}

	hash := sha256.Sum256([]byte(data))
	if !ecdsa.VerifyASN1(VerifyPublicKey, hash[:], sig) {
		return nil, errors.New("❌ 签名无效")
	}

	kv := make(map[string]string)
	r := bytes.NewReader(binary)

	info, err := buildinfo.Read(r)
	var binaryGoVersion string
	if err == nil {
		binaryGoVersion = info.GoVersion
		for _, s := range info.Settings {
			kv[s.Key] = s.Value
		}
	}

	// time
	t, err := time.Parse(time.RFC3339, release.BuildAt)
	if err != nil {
		return nil, errors.New("❌ 解析时间失败")
	}

	if share.BuildAtTime.After(t) && GetOption("go_allow_downgrade") != "1" {
		return nil, errors.New("❌ 版本已过期")
	}

	if kv["vcs.modified"] == "false" && binaryGoVersion == "go"+release.BuildVersion && strings.EqualFold(release.CommitHash, kv["vcs.revision"]) {
		uCtx := ghup.NewUpgrader(share.ReleaseRepo)

		uCtx.APIPrefix = share.ReleaseApiBase
		uCtx.Client = DefaultClient
		uCtx.Headers = UpgraderDefaultHeaders

		uCtx.Asset.Binary = binary
		uCtx.Asset.Size = len(binary) // needn't check
		uCtx.Asset.Hash = release.Sha256[kv["GOOS"]+"_"+kv["GOARCH"]]

		if err := uCtx.Verify(); err != nil {
			return nil, err
		}

		return uCtx, nil
	}

	return nil, errors.New("❌ 校验失败")
}
