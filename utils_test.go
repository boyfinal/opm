package opm

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var alphabet = "aaaaaaaaaaaaaaaaaaaaabcdefghijklmnopqrstuvwxyz"

// shortReader provides broken implementation of io.Reader for test
type shortReader struct{}

func (sr shortReader) Read(p []byte) (int, error) {
	return len(p) % 2, io.ErrUnexpectedEOF
}
func TestGenerateRandBytes(t *testing.T) {
	original := rand.Reader
	rand.Reader = shortReader{}
	defer func() {
		rand.Reader = original
	}()

	b, err := GenerateRandBytes(32)
	assert.NotNil(t, err, fmt.Sprintf("generateRandomBytes did not report a short read: only read %d bytes", len(b)))
}

func TestXOR(t *testing.T) {
	strs := []struct {
		a        []byte
		b        []byte
		expected []byte
	}{
		{[]byte("goodbye"), []byte("hello"), []byte{15, 10, 3, 8, 13}},
		{[]byte("onepuch"), []byte("onepiece"), []byte{0, 0, 0, 0, 28, 6, 11}},
		{nil, []byte("testing"), nil},
	}

	for _, v := range strs {
		if res := Xor(v.a, v.b); res != nil {
			if !bytes.Equal(res, v.expected) {
				t.Fatalf("XorBytes failed to return the expected result: got %v want %v", res, v.expected)
			}
		}
	}
}

func TestStrcon(t *testing.T) {
	ss := []string{"a", "b", "c", "d", "e"}
	expected := strings.Join(ss, "")
	res := Strcon(ss...)
	assert.Equal(t, res, expected, fmt.Sprintf("Strcon failed to concat %v: got %v want %v", strings.Join(ss, ","), res, expected))
}

func BenchmarkStrcon(b *testing.B) {
	ss := []string{"a", "b", "c", "d", "e"}
	for i := 0; i < b.N; i++ {
		s := Strcon(ss...)
		_ = s
	}
}

func TestContains(t *testing.T) {
	ss := []string{"a", "b", "c", "d", "e"}

	assert := assert.New(t)

	assert.Equal(Contains(ss, "a"), true, fmt.Sprintf("Strcon failed to contains 'a' in %v", ss))
	assert.NotEqual(Contains(ss, "f"), true, fmt.Sprintf("Strcon failed to contains 'f' in %v", ss))
}

func BenchmarkContains(b *testing.B) {
	ss := strings.Split(alphabet, "")
	s1 := "z"
	for i := 0; i < b.N; i++ {
		_ = Contains(ss, s1)
	}
}
func BenchmarkInArrayString(b *testing.B) {
	ss := strings.Split(alphabet, "")
	s1 := "z"
	for i := 0; i < b.N; i++ {
		_ = InArrayString(ss, s1)
	}
}

func TestNumFormat(t *testing.T) {
	intest := []struct {
		a        interface{}
		expected string
	}{
		{int(5), "5"},
		{int(-5), "-5"},
		{int8(5), "5"},
		{int8(-5), "-5"},
		{int16(5), "5"},
		{int16(-5), "-5"},
		{int32(555555555), "555555555"},
		{int32(-555555555), "-555555555"},
		{int64(5), "5"},
		{uint(5), "5"},
		{uint8(5), "5"},
		{uint16(5), "5"},
		{uint32(5), "5"},
		{uint64(5), "5"},
		{float32(5.03), "5.03"},
		{float32(-5.03), "-5.03"},
		{float64(5.5555), "5.5555"},
		{float64(-5.5555), "-5.5555"},
		{"5", "5"},
		{nil, ""},
	}

	for _, v := range intest {
		if res := NumFormat(v.a); res != v.expected {
			t.Fatalf("NumFormat failed to convert number %v to string: got %v want %v", v.a, res, v.expected)
		}
	}
}
