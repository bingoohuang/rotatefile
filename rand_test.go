package rotatefile

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"
	"unsafe"
)

// Implementations
// go test -bench . -benchmem
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go/22892986#22892986
/*
go test -bench . -benchmem
goos: darwin
goarch: amd64
pkg: github.com/bingoohuang/rotatefile
cpu: Intel(R) Core(TM) i7-8850H CPU @ 2.60GHz
BenchmarkRandomString-12                         5087995               222.0 ns/op            96 B/op          3 allocs/op
BenchmarkRandomBase64-12                        10229755               116.7 ns/op            64 B/op          3 allocs/op
BenchmarkRandomBase16-12                        12048391                99.89 ns/op           40 B/op          3 allocs/op
BenchmarkShortID-12                              5613924               212.5 ns/op            32 B/op          2 allocs/op
BenchmarkSecureRandomAlphaString-12              2526004               470.6 ns/op            64 B/op          3 allocs/op
BenchmarkSecureRandomString-12                   2371814               525.4 ns/op            61 B/op          3 allocs/op
BenchmarkBytesMaskImprRandReaderUnsafe-12        1000000              1023 ns/op             112 B/op          7 allocs/op
BenchmarkRunes-12                                2202200               463.1 ns/op            88 B/op          2 allocs/op
BenchmarkBytes-12                                3317440               375.6 ns/op            16 B/op          1 allocs/op
BenchmarkBytesRmndr-12                           4199571               285.7 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMask-12                            3281407               363.0 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImpr-12                        9648301               120.2 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprSrc-12                    11799037               101.0 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprSrcSB-12                  12178353                97.58 ns/op           16 B/op          1 allocs/op
BenchmarkBytesMaskImprSrcUnsafe-12              14530320                81.18 ns/op           16 B/op          1 allocs/op
PASS
ok      github.com/bingoohuang/rotatefile       21.982s
*/

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

func BenchmarkRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandomString(n)
	}
}

func RandomBase64(l int) string {
	buff := make([]byte, int(math.Ceil(float64(l)/1.33333333333)))
	rand.Read(buff)
	str := base64.RawURLEncoding.EncodeToString(buff)
	return str[:l] // strip 1 extra character we get from odd length results
}

func BenchmarkRandomBase64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandomBase64(n)
	}
}

func RandomBase16(l int) string {
	buff := make([]byte, int(math.Ceil(float64(l)/2)))
	rand.Read(buff)
	str := hex.EncodeToString(buff)
	return str[:l] // strip 1 extra character we get from odd length results
}

func BenchmarkRandomBase16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandomBase16(n)
	}
}

var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"

func ShortID(length int) string {
	ll := len(chars)
	b := make([]byte, length)
	rand.Read(b) // generates len(b) random bytes
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%ll]
	}
	return string(b)
}

func BenchmarkShortID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ShortID(n)
	}
}

func SecureRandomAlphaString(length int) string {
	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes = SecureRandomBytes(bufferSize)
		}
		if idx := int(randomBytes[j%length] & letterIdxMask); idx < len(letterBytes) {
			result[i] = letterBytes[idx]
			i++
		}
	}

	return string(result)
}

func BenchmarkSecureRandomAlphaString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SecureRandomAlphaString(n)
	}
}

// SecureRandomString returns a string of the requested length,
// made from the byte characters provided (only ASCII allowed).
// Uses crypto/rand for security. Will panic if len(availableCharBytes) > 256.
func SecureRandomString(availableCharBytes string, length int) string {

	// Compute bitMask
	availableCharLength := len(availableCharBytes)
	if availableCharLength == 0 || availableCharLength > 256 {
		panic("availableCharBytes length must be greater than 0 and less than or equal to 256")
	}
	var bitLength byte
	var bitMask byte
	for bits := availableCharLength - 1; bits != 0; {
		bits = bits >> 1
		bitLength++
	}
	bitMask = 1<<bitLength - 1

	// Compute bufferSize
	bufferSize := length + length/3

	// Create random string
	result := make([]byte, length)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			// Random byte buffer is empty, get a new one
			randomBytes = SecureRandomBytes(bufferSize)
		}
		// Mask bytes to get an index into the character slice
		if idx := int(randomBytes[j%length] & bitMask); idx < availableCharLength {
			result[i] = availableCharBytes[idx]
			i++
		}
	}

	return string(result)
}

func BenchmarkSecureRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SecureRandomString(letterBytes, n)
	}
}

func RandStringBytesMaskImprRandReaderUnsafe(length uint) (string, error) {
	const (
		charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		charIdxBits = 6                  // 6 bits to represent a letter index
		charIdxMask = 1<<charIdxBits - 1 // All 1-bits, as many as charIdxBits
		charIdxMax  = 63 / charIdxBits   // # of letter indices fitting in 63 bits
	)

	buffer := make([]byte, length)
	charsetLength := len(charset)
	max := big.NewInt(int64(1 << uint64(charsetLength)))

	limit, err := crand.Int(crand.Reader, max)
	if err != nil {
		return "", err
	}

	for index, cache, remain := int(length-1), limit.Int64(), charIdxMax; index >= 0; {
		if remain == 0 {
			limit, err = crand.Int(crand.Reader, max)
			if err != nil {
				return "", err
			}

			cache, remain = limit.Int64(), charIdxMax
		}

		if idx := int(cache & charIdxMask); idx < charsetLength {
			buffer[index] = charset[idx]
			index--
		}

		cache >>= charIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&buffer)), nil
}

func BenchmarkBytesMaskImprRandReaderUnsafe(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	const length = 16

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandStringBytesMaskImprRandReaderUnsafe(length)
		}
	})
}

// SecureRandomBytes returns the requested number of bytes using crypto/rand
func SecureRandomBytes(length int) []byte {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatal("Unable to generate random bytes")
	}
	return randomBytes
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func RandStringBytesMask(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

func RandStringBytesMaskImpr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func RandStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// Benchmark functions

const n = 16

func BenchmarkRunes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringRunes(n)
	}
}

func BenchmarkBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringBytes(n)
	}
}

func BenchmarkBytesRmndr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringBytesRmndr(n)
	}
}

func BenchmarkBytesMask(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringBytesMask(n)
	}
}

func BenchmarkBytesMaskImpr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringBytesMaskImpr(n)
	}
}

func BenchmarkBytesMaskImprSrc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringBytesMaskImprSrc(n)
	}
}
func BenchmarkBytesMaskImprSrcSB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringBytesMaskImprSrcSB(n)
	}
}

func BenchmarkBytesMaskImprSrcUnsafe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringBytesMaskImprSrcUnsafe(n)
	}
}
