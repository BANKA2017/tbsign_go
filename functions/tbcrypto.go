package _function

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"strings"

	"github.com/OneOfOne/xxhash"
)

// https://github.com/lumina37/aiotieba/tree/master/aiotieba/helper/crypto/src/tbcrypto
// https://github.com/BaWuZhuShou/AioTieba4DotNet/blob/master/AioTieba4DotNet/Internal/TbCrypto.cs
// translate by copilot(Claude Haiku 4.5)

// Constants
const (
	TbcAndroidIdSize = 16
	TbcMd5HashSize   = 16
	TbcMd5StrSize    = TbcMd5HashSize * 2
	TbcSha1HashSize  = 20
	TbcSha1HexSize   = TbcSha1HashSize * 2
	HasherNum        = 4
	StepSize         = 5
	HashSizeInBit    = 32
)

var (
	Cuid2Prefix       = []byte("com.baidu")
	Cuid3Prefix       = []byte("com.helios")
	HexUppercaseTable = "0123456789ABCDEF"
)

// TbCrypto holds Android device identifiers
type TbCrypto struct {
	AndroidID string
	UUID      string
}

// NewTbCrypto creates a new TbCrypto instance
func NewTbCrypto(androidID, uuid string) (*TbCrypto, error) {
	if len(androidID) != 16 {
		return nil, fmt.Errorf("invalid size of android_id. Expected 16, got %d", len(androidID))
	}
	if len(uuid) != 36 {
		return nil, fmt.Errorf("invalid size of uuid. Expected 36, got %d", len(uuid))
	}

	return &TbCrypto{
		AndroidID: androidID,
		UUID:      uuid,
	}, nil
}

// TbcSha1Base32Size calculates the Base32 length for SHA1 hash
func TbcSha1Base32Size() int {
	return base32Len(TbcSha1HashSize)
}

// base32Len calculates Base32 encoded length
func base32Len(len int) int {
	result := (len / 5) * 8
	if len%5 != 0 {
		result += 8
	}
	return result
}

// tbcUpdate updates the hash internal state
func tbcUpdate(sec uint64, hashVal uint64, start uint64, flag bool) uint64 {
	end := start + HashSizeInBit
	secTemp := sec
	var9 := (uint64(1) << end) - 1
	var5 := (var9 & sec) >> start

	if flag {
		var5 ^= hashVal
	} else {
		var5 &= hashVal
	}

	for i := range uint64(HashSizeInBit) {
		opIdx := start + i
		if (var5 & (uint64(1) << i)) != 0 {
			secTemp |= (uint64(1) << opIdx)
		} else {
			secTemp &= ^(uint64(1) << opIdx)
		}
	}

	return secTemp
}

// tbcWriteBuffer writes internal state to buffer
func tbcWriteBuffer(sec uint64) []byte {
	buffer := make([]byte, StepSize)
	tmpSec := sec
	for i := range StepSize {
		buffer[i] = byte(tmpSec & 0xFF)
		tmpSec >>= 8
	}
	return buffer
}

func tbcHeliosHash(src []byte, size int) []byte {
	sec := (uint64(1) << 40) - 1
	buffer := make([]byte, HasherNum*StepSize)
	for i := range StepSize {
		buffer[i] = 0xFF
	}

	// First hash using CRC32
	crc32Hash := crc32.NewIEEE()
	crc32Hash.Write(src[:size])
	crc32Hash.Write(buffer[:StepSize])
	crc32Val := crc32Hash.Sum32()

	sec = tbcUpdate(sec, uint64(crc32Val), 8, false)

	buf1 := tbcWriteBuffer(sec)
	copy(buffer[StepSize:StepSize*2], buf1)
	// Second hash using XXHash32
	xxhash32 := xxhash.New32()
	xxhash32.Write(src[:size])
	xxhash32.Write(buffer[:StepSize*2])
	xxhashVal32 := xxhash32.Sum32()

	sec = tbcUpdate(sec, uint64(xxhashVal32), 0, true)

	buf2 := tbcWriteBuffer(sec)
	copy(buffer[StepSize*2:StepSize*3], buf2)
	// Third hash using XXHash32
	xxhash32.Write(buffer[StepSize*2 : StepSize*3])
	xxhashVal32 = xxhash32.Sum32()

	sec = tbcUpdate(sec, uint64(xxhashVal32), 1, true)

	buf3 := tbcWriteBuffer(sec)
	copy(buffer[StepSize*3:StepSize*4], buf3)

	// Fourth hash using CRC32
	crc32Hash.Write(buffer[StepSize : StepSize*4])
	crc32Val = crc32Hash.Sum32()

	sec = tbcUpdate(sec, uint64(crc32Val), 7, true)
	// Return final result
	result := tbcWriteBuffer(sec)

	return result
}

// tbcCuidGalaxy2 generates CUID Galaxy 2
func (tc *TbCrypto) tbcCuidGalaxy2(androidID []byte) string {
	// Build MD5 input buffer
	md5Buffer := make([]byte, len(Cuid2Prefix)+TbcAndroidIdSize)
	copy(md5Buffer, Cuid2Prefix)
	copy(md5Buffer[len(Cuid2Prefix):], androidID)

	// Calculate MD5
	md5Hash := md5.Sum(md5Buffer)

	// Convert to hex string
	sb := strings.ToUpper(hex.EncodeToString(md5Hash[:]))
	sb += "|V"

	// Calculate Helios hash
	heliosHash := tbcHeliosHash([]byte(sb), TbcMd5StrSize)

	// Base32 encode and append
	sb += base32Encode(heliosHash)

	return sb
}

// tbcC3Aid generates CUID 3 AID
func (tc *TbCrypto) tbcC3Aid(androidID []byte, uuid []byte) string {
	sha1InputLength := len(Cuid3Prefix) + TbcAndroidIdSize + len(uuid)
	dstOffset := 5 + TbcSha1Base32Size()

	// Build SHA1 input buffer
	sha1Buffer := make([]byte, sha1InputLength)
	copy(sha1Buffer, Cuid3Prefix)
	copy(sha1Buffer[len(Cuid3Prefix):], androidID)
	copy(sha1Buffer[len(Cuid3Prefix)+TbcAndroidIdSize:], uuid)

	// Calculate SHA1
	sha1Hash := sha1.Sum(sha1Buffer)

	// Build string with Base32 encoding of SHA1
	sb := "A00-" + base32Encode(sha1Hash[:])
	sb += "-"

	// Calculate Helios hash
	heliosHash := tbcHeliosHash([]byte(sb), dstOffset)

	// Base32 encode and append
	sb += base32Encode(heliosHash)

	return sb
}

// base32Encode encodes bytes to Base32 string
func base32Encode(input []byte) string {
	const base32Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

	sb := strings.Builder{}
	bitBuffer := 0
	bitCount := 0

	for _, b := range input {
		bitBuffer = (bitBuffer << 8) | int(b)
		bitCount += 8

		for bitCount >= 5 {
			index := (bitBuffer >> (bitCount - 5)) & 31
			sb.WriteByte(base32Chars[index])
			bitCount -= 5
		}
	}

	if bitCount > 0 {
		tailIndex := (bitBuffer << (5 - bitCount)) & 31
		sb.WriteByte(base32Chars[tailIndex])
	}

	return sb.String()
}

// CuidGalaxy2 generates CUID Galaxy 2 from device identifiers
func (tc *TbCrypto) CuidGalaxy2() (string, error) {
	androidIDBytes := []byte(tc.AndroidID)
	if len(androidIDBytes) != TbcAndroidIdSize {
		return "", fmt.Errorf("invalid size of android_id. Expected %d, got %d", TbcAndroidIdSize, len(androidIDBytes))
	}

	return tc.tbcCuidGalaxy2(androidIDBytes), nil
}

// C3Aid generates CUID 3 AID from device identifiers
func (tc *TbCrypto) C3Aid() (string, error) {
	androidIDBytes := []byte(tc.AndroidID)
	uuidBytes := []byte(tc.UUID)

	if len(androidIDBytes) != TbcAndroidIdSize {
		return "", fmt.Errorf("invalid size of android_id. Expected %d, got %d", TbcAndroidIdSize, len(androidIDBytes))
	}

	if len(uuidBytes) != 36 {
		return "", fmt.Errorf("invalid size of uuid. Expected 36, got %d", len(uuidBytes))
	}

	return tc.tbcC3Aid(androidIDBytes, uuidBytes), nil
}
