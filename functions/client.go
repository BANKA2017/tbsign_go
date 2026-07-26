package _function

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/pbkdf2"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/OneOfOne/xxhash"
	"github.com/google/uuid"
)

type Client struct {
	Account *_type.TypeCookie

	// for client
	AndroidID string
	UUID      string

	ClientID string
	SampleID string

	cuid        string
	cuidGalaxy2 string
	c3Aid       string

	aesCBCSecKey []byte
	aesCBCCipher cipher.Block

	aesECBSecretKey []byte
	aesECBCipher    cipher.Block

	Zid string

	mu sync.Mutex
}

func (tc *Client) RandomAndroidID() string {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	androidID, _ := RandomTokenBuilder(8)

	tc.AndroidID = hex.EncodeToString(androidID)

	// reset c3aid and cuid
	tc.c3Aid = ""
	tc.cuidGalaxy2 = ""

	return tc.AndroidID
}

func (tc *Client) RandomUUID() string {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.UUID = uuid.New().String()

	// reset c3aid and cuid
	tc.c3Aid = ""
	tc.cuidGalaxy2 = ""
	tc.cuid = ""

	return tc.UUID
}

// https://github.com/lumina37/aiotieba/tree/master/aiotieba/helper/crypto/src/tbcrypto
// https://github.com/BaWuZhuShou/AioTieba4DotNet/blob/master/AioTieba4DotNet/Internal/TbCrypto.cs
// translate by copilot(Claude Haiku 4.5)

// Constants
const (
	tbcAndroidIdSize = 16
	tbcMd5HashSize   = 16
	tbcMd5StrSize    = tbcMd5HashSize * 2
	tbcSha1HashSize  = 20
	tbcSha1HexSize   = tbcSha1HashSize * 2
	hasherNum        = 4
	stepSize         = 5
	hashSizeInBit    = 32
)

var (
	cuid2Prefix   = "com.baidu"
	cuid3Prefix   = "com.helios"
	base32Encoder = base32.StdEncoding.WithPadding(base32.NoPadding)
)

// TbCrypto holds Android device identifiers

// // base32Len calculates Base32 encoded length
// func base32Len(len int) int {
// 	result := (len / 5) * 8
// 	if len%5 != 0 {
// 		result += 8
// 	}
// 	return result
// }

// tbcUpdate updates the hash internal state
func tbcUpdate(sec uint64, hashVal uint64, start uint64, flag bool) uint64 {
	end := start + hashSizeInBit
	secTemp := sec
	var9 := (uint64(1) << end) - 1
	var5 := (var9 & sec) >> start

	if flag {
		var5 ^= hashVal
	} else {
		var5 &= hashVal
	}

	for i := range uint64(hashSizeInBit) {
		opIdx := start + i
		if (var5 & (uint64(1) << i)) != 0 {
			secTemp |= uint64(1) << opIdx
		} else {
			secTemp &= ^(uint64(1) << opIdx)
		}
	}

	return secTemp
}

// heliosHashWriteBuffer writes internal state to buffer
func heliosHashWriteBuffer(sec uint64) []byte {
	buffer := make([]byte, stepSize)
	tmpSec := sec
	for i := range stepSize {
		buffer[i] = byte(tmpSec & 0xFF)
		tmpSec >>= 8
	}
	return buffer
}

func heliosHash(src []byte, size int) []byte {
	sec := (uint64(1) << 40) - 1
	buffer := make([]byte, hasherNum*stepSize)
	for i := range stepSize {
		buffer[i] = 0xFF
	}

	// First hash using CRC32
	crc32Hash := crc32.NewIEEE()
	crc32Hash.Write(src[:size])
	crc32Hash.Write(buffer[:stepSize])
	crc32Val := crc32Hash.Sum32()

	sec = tbcUpdate(sec, uint64(crc32Val), 8, false)

	buf1 := heliosHashWriteBuffer(sec)
	copy(buffer[stepSize:stepSize*2], buf1)
	// Second hash using XXHash32
	xxhash32 := xxhash.New32()
	xxhash32.Write(src[:size])
	xxhash32.Write(buffer[:stepSize*2])
	xxhashVal32 := xxhash32.Sum32()

	sec = tbcUpdate(sec, uint64(xxhashVal32), 0, true)

	buf2 := heliosHashWriteBuffer(sec)
	copy(buffer[stepSize*2:stepSize*3], buf2)
	// Third hash using XXHash32
	xxhash32.Write(buffer[stepSize*2 : stepSize*3])
	xxhashVal32 = xxhash32.Sum32()

	sec = tbcUpdate(sec, uint64(xxhashVal32), 1, true)

	buf3 := heliosHashWriteBuffer(sec)
	copy(buffer[stepSize*3:stepSize*4], buf3)

	// Fourth hash using CRC32
	crc32Hash.Write(buffer[stepSize : stepSize*4])
	crc32Val = crc32Hash.Sum32()

	sec = tbcUpdate(sec, uint64(crc32Val), 7, true)
	// Return final result
	result := heliosHashWriteBuffer(sec)

	return result
}

// Cuid only for legacy versions like 9.x, >= 11.x use cuid_galaxy2
func (tc *Client) Cuid() string {
	if tc.cuid == "" {
		tc.cuid = "baidutiebaapp" + tc.UUID
	}
	return tc.cuid
}

// CuidGalaxy2 generates CUID Galaxy 2 from device identifiers
func (tc *Client) CuidGalaxy2() (string, error) {
	if tc.cuidGalaxy2 == "" {
		if len(tc.AndroidID) != tbcAndroidIdSize {
			return "", errors.New("invalid android id")
		}

		// Convert to hex string
		sb := strings.ToUpper(Md5(cuid2Prefix + tc.AndroidID))
		sb += "|V"

		// Calculate Helios hash
		heliosHash := heliosHash([]byte(sb), tbcMd5StrSize)

		// Base32 encode and append
		sb += base32Encoder.EncodeToString(heliosHash)

		tc.cuidGalaxy2 = sb

	}

	return tc.cuidGalaxy2, nil
}

// C3Aid generates CUID 3 AID from device identifiers
func (tc *Client) C3Aid() (string, error) {
	if tc.c3Aid == "" {
		if len(tc.AndroidID) != tbcAndroidIdSize {
			return "", errors.New("invalid android id")
		} else if len(tc.UUID) != 36 {
			return "", errors.New("invalid uuid")
		}

		dstOffset := 5 + 32 // base32Len(tbcSha1HashSize)
		// Calculate SHA1
		sha1Hash := sha1.Sum([]byte(cuid3Prefix + tc.AndroidID + tc.UUID))

		// Build string with Base32 encoding of SHA1
		sb := "A00-" + base32Encoder.EncodeToString(sha1Hash[:])
		sb += "-"

		// Calculate Helios hash
		heliosHash := heliosHash([]byte(sb), dstOffset)

		// Base32 encode and append
		sb += base32Encoder.EncodeToString(heliosHash)

		tc.c3Aid = sb
	}

	return tc.c3Aid, nil
}

func (tc *Client) AESCBCSecKey() ([]byte, error) {
	if tc.aesCBCSecKey == nil {
		// NEVER CREATE DEADLOCK!!!
		// tc.mu.Lock()
		// defer tc.mu.Unlock()

		key, err := RandomTokenBuilder(16)
		if err != nil {
			return nil, fmt.Errorf("generate random key error: %w", err)
		}
		tc.aesCBCSecKey = key
	}
	return tc.aesCBCSecKey, nil
}

func (tc *Client) AESCBCCipher() (cipher.Block, error) {
	if tc.aesCBCCipher == nil {
		tc.mu.Lock()
		defer tc.mu.Unlock()

		key, err := tc.AESCBCSecKey()
		if err != nil {
			return nil, err
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("create aes cipher error: %w", err)
		}

		tc.aesCBCCipher = block
	}
	return tc.aesCBCCipher, nil
}

func (tc *Client) AESECBSecKey() ([]byte, error) {
	if tc.aesECBSecretKey == nil {
		// NEVER CREATE DEADLOCK!!!
		// tc.mu.Lock()
		// defer tc.mu.Unlock()

		key, err := RandomTokenBuilder(31)
		if err != nil {
			return nil, fmt.Errorf("generate random key error: %w", err)
		}

		tc.aesECBSecretKey = key
	}

	return tc.aesECBSecretKey, nil
}

func (tc *Client) AESECBCipher() (cipher.Block, error) {
	if tc.aesECBCipher == nil {
		tc.mu.Lock()
		defer tc.mu.Unlock()

		// double check
		if tc.aesECBCipher != nil {
			return tc.aesECBCipher, nil
		}

		key, err := tc.AESECBSecKey()
		if err != nil {
			return nil, err
		}

		salt := []byte{
			0xa4, 0x0b, 0xc8, 0x34,
			0xd6, 0x95, 0xf3, 0x13,
		}

		wsSecretKey, _ := pbkdf2.Key(
			sha1.New,
			string(key),
			salt,
			5,
			32,
		)

		block, err := aes.NewCipher(wsSecretKey)
		if err != nil {
			return nil, fmt.Errorf("create aes ecb cipher error: %w", err)
		}

		tc.aesECBCipher = block
	}

	return tc.aesECBCipher, nil
}

// https://github.com/lumina37/aiotieba/blob/master/aiotieba/api/init_z_id/_api.py
// https://github.com/BaWuZhuShou/AioTieba4DotNet/blob/master/AioTieba4DotNet/Api/InitZId/InitZId.cs
// translate by copilot(Claude Haiku 4.5)

const (
	zidAppKey = "200033"
	zidSecKey = "ea737e4f435b53786043369d2e5ace4f"
)

type zidParamsModuleSection struct {
	Zid string `json:"zid"`
}

type zidParams struct {
	ModuleSection []zidParamsModuleSection `json:"module_section"`
}

type zidRes struct {
	Data      string `json:"data"`
	RequestID int64  `json:"request_id"`
	Skey      string `json:"skey"`
	T         int    `json:"t"`
}

type zidTokenDecrypted struct {
	Token string `json:"token"`
}

func (tc *Client) SyncZid() (string, error) {
	cbcSecKey, err := tc.AESCBCSecKey()
	if err != nil {
		return "", fmt.Errorf("get aes cbc sec key error: %w", err)
	}

	cbcCipher, err := tc.AESCBCCipher()
	if err != nil {
		return "", fmt.Errorf("get aes cbc cipher error: %w", err)
	}

	if len(tc.AndroidID) != tbcAndroidIdSize {
		return "", errors.New("invalid android id")
	} else if len(tc.UUID) != 36 {
		return "", errors.New("invalid uuid")
	}

	now := strconv.FormatInt(time.Now().Unix(), 10)

	xyusMD5FinalStr := Md5(Md5(tc.AndroidID+tc.UUID) + "|0")

	params := zidParams{
		ModuleSection: []zidParamsModuleSection{
			{Zid: xyusMD5FinalStr},
		},
	}

	reqBodyJSON, err := JsonEncode(params)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %w", err)
	}

	var gzipBuf bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&gzipBuf, gzip.DefaultCompression)
	if err != nil {
		return "", fmt.Errorf("gzip writer error: %w", err)
	}
	if _, err := gzipWriter.Write(reqBodyJSON); err != nil {
		return "", fmt.Errorf("gzip write error: %w", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		return "", err
	}
	reqBodyCompressed := gzipBuf.Bytes()

	paddedData := pkcs7Pad(reqBodyCompressed, aes.BlockSize)

	// cbc
	iv := make([]byte, 16)
	blockMode := cipher.NewCBCEncrypter(cbcCipher, iv)
	reqBodyAES := make([]byte, len(paddedData))
	blockMode.CryptBlocks(reqBodyAES, paddedData)

	reqBodyMD5 := md5.Sum(reqBodyCompressed)
	reqData := append(reqBodyAES, reqBodyMD5[:]...)

	pathCombineMD5Str := Md5(zidAppKey + now + zidSecKey)

	reqQuerySKeyB64 := base64.StdEncoding.EncodeToString(Rc442(xyusMD5FinalStr, cbcSecKey))

	fullURL := fmt.Sprintf("https://sofire.baidu.com/c/11/z/100/%s/%s/%s", zidAppKey, now, pathCombineMD5Str) + "?skey=" + url.QueryEscape(reqQuerySKeyB64)

	headers := map[string]string{
		"x-device-id": xyusMD5FinalStr,
		"User-Agent":  fmt.Sprintf("x6/%s/%s/4.4.1.3", zidAppKey, ClientVersion),
		"x-plu-ver":   "x6/4.4.1.3",
	}

	bodyBytes, err := TBFetch(fullURL, http.MethodPost, reqData, headers)

	if err != nil {
		return "", fmt.Errorf("read response error: %w", err)
	}

	var resJSON zidRes
	if err := JsonDecode(bodyBytes, &resJSON); err != nil {
		return "", fmt.Errorf("json unmarshal error: %w", err)
	}

	if resJSON.Skey == "" {
		return "", fmt.Errorf("skey not found in response")
	}

	resQuerySkey, err := base64.StdEncoding.DecodeString(resJSON.Skey)
	if err != nil {
		return "", fmt.Errorf("decode skey error: %w", err)
	}

	resAESSec := Rc442(xyusMD5FinalStr, resQuerySkey)

	if resJSON.Data == "" {
		return "", fmt.Errorf("data not found in response")
	}

	resData, err := base64.StdEncoding.DecodeString(resJSON.Data)
	if err != nil {
		return "", fmt.Errorf("decode data error: %w", err)
	}

	// AES-CBC decrypt
	ivDec := make([]byte, 16)

	aesCipher, err := aes.NewCipher(resAESSec)
	if err != nil {
		return "", fmt.Errorf("create aes cipher for decrypt error: %w", err)
	}

	blockModeDec := cipher.NewCBCDecrypter(aesCipher, ivDec)
	decryptedData := make([]byte, len(resData))
	blockModeDec.CryptBlocks(decryptedData, resData)

	// remove md5 suffix
	if len(decryptedData) > 16 {
		decryptedData = decryptedData[:len(decryptedData)-16]
	}

	// remove PKCS7 padding
	decryptedData = pkcs7Unpad(decryptedData)

	var resultData zidTokenDecrypted
	if err := JsonDecode(decryptedData, &resultData); err != nil {
		return "", fmt.Errorf("json unmarshal result error: %w", err)
	}

	zid := resultData.Token
	if resultData.Token == "" {
		return "", fmt.Errorf("token not found in result data")
	}

	tc.Zid = zid

	return tc.Zid, nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	length := len(data)
	unpadding := int(data[length-1])

	if unpadding > length {
		return data
	}

	return data[:(length - unpadding)]
}

// Rc442 RC4加密变体，每字节额外XOR 42
func Rc442(xyusMd5Str string, cbcSecKey []byte) []byte {
	keyBytes := []byte(xyusMd5Str)

	m := make([]byte, 256)
	for i := range 256 {
		m[i] = byte(i)
	}

	j := 0
	for i := range 256 {
		j = (j + int(m[i]) + int(keyBytes[i%len(keyBytes)])) & 0xFF
		m[i], m[j] = m[j], m[i]
	}

	x, y := 0, 0
	dst := make([]byte, len(cbcSecKey))

	for i := range cbcSecKey {
		x = (x + 1) & 0xFF
		a := int(m[x])
		y = (y + a) & 0xFF
		b := int(m[y])

		m[x], m[y] = m[y], m[x]

		// 关键：输出字节XOR上m[(a+b)%256]，然后再XOR 42
		kgenerator := m[(a+b)&0xFF]
		dst[i] = cbcSecKey[i] ^ kgenerator ^ 42
	}

	return dst
}

// bb64
// by chatgpt && jetbrains ai
var BB64Alphabet = []byte("qogjOuCRNkfil5p4SQ3LAmxGKZTdesvB6z_YPahMI9t80rJyHW1DEwFbc7nUVX2-")

func EnuidEncode(cuidGalaxy2 string) (string, error) {
	if cuidGalaxy2 == "" {
		return "", errors.New("cuidGalaxy2 is empty")
	}

	source := []byte(cuidGalaxy2)
	working := make([]byte, ((len(source)+2)/3)*4+2)
	copy(working, source)
	encodedLength := EncodeInPlace(working, len(source), 0)

	return string(working[:encodedLength-1]), nil
}

func EncodeInPlace(buffer []byte, inputLength int, mode int) int {
	if inputLength == 0 {
		return -1
	}

	remainder := inputLength % 3
	blockCount := inputLength / 3
	prefixLength := 4 * blockCount
	table := BuildTable(mode, false)
	XorWords(table.State, buffer, inputLength)
	temp := make([]byte, 4)
	if remainder > 0 {
		copy(temp, buffer[3*blockCount:])
		EncodeTriplet(table.Map, temp, buffer, 3*blockCount+blockCount)
	}

	outputLength := prefixLength
	if remainder > 0 {
		outputLength += 4
	}

	for i := range blockCount {
		sourceIndex := 3 * (blockCount - 1 - i)
		destinationIndex := prefixLength - 4*(i+1)
		EncodeTriplet(table.Map, buffer[sourceIndex:sourceIndex+3], buffer, destinationIndex)
	}

	buffer[outputLength] = byte(remainder + 65)
	buffer[outputLength+1] = 0

	return outputLength + 2
}

type BB64Table struct {
	State []int32
	Map   []byte
}

func BuildTable(mode int, fillLookup bool) BB64Table {
	state := make([]int32, 49)
	mp := make([]byte, 196)
	offset := mode & 0x3f
	tailLength := 64 - offset

	copy(mp[132:132+tailLength], BB64Alphabet[offset:])
	copy(mp[132+tailLength:132+64], BB64Alphabet[:offset])
	rotated := int32(uint32(mode)<<27 | uint32(mode)>>5)
	state[0] = rotated ^ int32(758653732<<uint32(rotated&0xf))

	if fillLookup {
		for i := range 128 {
			mp[i+4] = 64
		}
		for i := range 64 {
			mp[mp[i+132]+4] = byte(i)
		}
	}

	return BB64Table{
		State: state,
		Map:   mp,
	}
}

func XorWords(state []int32, buffer []byte, length int) {
	seed := state[0]
	wholeWords := length >> 2

	for i := range wholeWords {
		offset := i * 4

		word := int32(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
		word = word ^ int32(uint32(word)>>3|uint32(word)<<29)

		binary.LittleEndian.PutUint32(buffer[offset:offset+4], uint32(word))

	}

	if wholeWords < (length+3)>>2 {
		var seedBytes [4]byte
		binary.LittleEndian.PutUint32(seedBytes[:], uint32(seed))

		start := wholeWords * 4
		for i := start; i < length; i++ {
			buffer[i] ^= seedBytes[i-start]
		}
	}
}

func EncodeTriplet(mp []byte, input []byte, destination []byte, destinationIndex int) {
	b0 := input[0]
	b1 := input[1]
	b2 := input[2]

	destination[destinationIndex] = mp[132+int(b0&0x3f)]
	destination[destinationIndex+1] = mp[132+int(b1&0x3f)]
	destination[destinationIndex+2] = mp[132+int(b2&0x3f)]
	destination[destinationIndex+3] = mp[132+int(((b1>>6)*4)|((b0>>6)*16)|(b2>>6))]
}
