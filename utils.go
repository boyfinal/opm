package opm

import (
	"bytes"
	"crypto/rand"
	"strconv"
	"strings"
)

func GenerateRandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func XorBytes(a, b []byte) []byte {
	n := len(a)
	if n > len(b) {
		n = len(b)
	}

	res := make([]byte, n)
	for i := 0; i < n; i++ {
		res[i] = a[i] ^ b[i]
	}

	return res
}

func Strcon(ss ...string) string {
	var b bytes.Buffer
	for _, s := range ss {
		b.WriteString(s)
	}
	return b.String()
}

func Contains(vals []string, s string) bool {
	for _, v := range vals {
		if v == s {
			return true
		}
	}

	return false
}

func NumFormat(v interface{}) string {
	switch s := v.(type) {
	case int:
		return numFormart(int64(s))
	case int8:
		return numFormart(int64(s))
	case int16:
		return numFormart(int64(s))
	case int32:
		return numFormart(int64(s))
	case int64:
		return numFormart(int64(s))
	case uint:
		return numFormart(int64(s))
	case uint8:
		return numFormart(int64(s))
	case uint16:
		return numFormart(int64(s))
	case uint32:
		return numFormart(int64(s))
	case uint64:
		return numFormart(int64(s))
	case float32:
		return floatFormart(float64(s))
	case float64:
		return floatFormart(s)
	case string:
		return v.(string)
	default:
		return ""
	}
}

func numFormart(v int64) string {
	s := strconv.FormatInt(v, 10)
	return dformat(s, true)
}

func dformat(s string, t bool) string {
	if s == "" {
		return s
	}

	if s[0:1] == "-" {
		return "-" + dformat(s[1:], t)
	}

	var runes []rune
	if t {
		s = reverse(s)
	}

	for i, r := range s {
		runes = append(runes, r)
		if (i+1)%3 == 0 {
			runes = append(runes, 44)
		}
	}

	if i := len(runes) - 1; runes[i] == 44 {
		runes = runes[:i]
	}

	if t {
		return runereverse(runes)
	}

	return string(runes)
}

func floatFormart(v float64) string {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	ar := strings.Split(s, ".")

	ss := dformat(ar[0], true)
	if len(ar) > 1 {
		ss = ss + "." + dformat(ar[1], false)
	}

	return ss
}

func reverse(s string) string {
	n := len(s)
	runes := make([]rune, n)
	for i, r := range s {
		runes[n-i-1] = r
	}

	return string(runes)
}

func runereverse(r []rune) string {
	n := len(r)
	runes := make([]rune, n)
	for i, v := range r {
		runes[n-i-1] = v
	}

	return string(runes)
}
