package opm

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"strconv"
	"strings"
)

// GenerateRandBytes returns generated random bytes
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

func InArrayString(arr []string, in string) bool {
	n := len(arr)
	if n%2 == 1 && in == arr[n-1] {
		return true
	}

	k := n / 2
	for i := 0; i < k; i++ {
		if in == arr[i] {
			return true
		}

		if in == arr[i+k] {
			return true
		}
	}

	return false
}

func InArrayString1(arr []string, in string) bool {
	n := len(arr)
	for i := 0; i < n; i++ {
		if in == arr[i] {
			return true
		}
	}

	return false
}

func InArrayString2(arr []string, in string) bool {
	for _, v := range arr {
		if v == in {
			return true
		}
	}

	return false
}

func Contains(arr interface{}, in interface{}) bool {
	if kind := reflect.TypeOf(arr).Kind(); kind != reflect.Slice && kind != reflect.Array {
		return false
	}

	values := reflect.ValueOf(arr)
	for i := 0; i < values.Len(); i++ {
		if reflect.DeepEqual(in, values.Index(i).Interface()) {
			return true
		}
	}

	return false
}

func NumFormat(v interface{}) string {
	switch s := v.(type) {
	case int:
		return strconv.Itoa(s)
	case int8:
		return strconv.FormatInt(int64(s), 10)
	case int16:
		return strconv.FormatInt(int64(s), 10)
	case int32:
		return strconv.FormatInt(int64(s), 10)
	case int64:
		return strconv.FormatInt(int64(s), 10)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case uint8:
		return strconv.FormatUint(uint64(s), 10)
	case uint16:
		return strconv.FormatUint(uint64(s), 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case uint64:
		return strconv.FormatUint(uint64(s), 10)
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
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

	println(s, ar)

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
