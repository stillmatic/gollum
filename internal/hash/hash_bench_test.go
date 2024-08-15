package hash

// copied from https://gist.github.com/wizjin/e103e1040db0c4c427db4104cce67566

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"github.com/cespare/xxhash/v2"
	"hash"
	"hash/fnv"
	"testing"
)

// NB chua: we care about caching / hashing lots of tokens.
// That probably starts mattering at O(1k), O(10k), O(100k) tokens.
// note that there are ~4 characters per token, and 1-4 bytes per character in UTF-8
// so 1k tokens is 4k-16k bytes, 10k tokens is 40k-160k bytes, 100k tokens is 400k-1.6M bytes.
const (
	K       = 1024
	DATALEN = 512 * K
)

func runHash(b *testing.B, h hash.Hash, n int) {
	var data = make([]byte, n)
	rand.Read(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Write(data)
		h.Sum(nil)
	}
}

func BenchmarkFNV32(b *testing.B) {
	runHash(b, fnv.New32(), DATALEN)
}

func BenchmarkFNV64(b *testing.B) {
	runHash(b, fnv.New64(), DATALEN)
}

func BenchmarkFNV128(b *testing.B) {
	runHash(b, fnv.New128(), DATALEN)
}

func BenchmarkMD5(b *testing.B) {
	runHash(b, md5.New(), DATALEN)
}

func BenchmarkSHA1(b *testing.B) {
	runHash(b, sha1.New(), DATALEN)
}

func BenchmarkSHA224(b *testing.B) {
	runHash(b, sha256.New224(), DATALEN)
}

func BenchmarkSHA256(b *testing.B) {
	runHash(b, sha256.New(), DATALEN)
}

func BenchmarkSHA512(b *testing.B) {
	runHash(b, sha512.New(), DATALEN)
}

//	func BenchmarkMurmur3(b *testing.B) {
//		runHash(b, murmur3.New32(), DATALEN)
//	}
func BenchmarkXxhash(b *testing.B) {
	runHash(b, xxhash.New(), DATALEN)
}
