package opm

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"strconv"
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

// Xor []byte
func Xor(a, b []byte) []byte {
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

// Strcon concat multiple string
func Strcon(ss ...string) string {
	var b bytes.Buffer
	for _, s := range ss {
		b.WriteString(s)
	}
	return b.String()
}

// func InArrayString(arr []string, in string) bool {
// 	n := len(arr)
// 	if n%2 == 1 && in == arr[n-1] {
// 		return true
// 	}

// 	k := n / 2
// 	for i := 0; i < k; i++ {
// 		if in == arr[i] || in == arr[i+k] {
// 			return true
// 		}
// 	}

// 	return false
// }

// InArrayString determines whether an array string includes a certain value among its entries
func InArrayString(arr []string, in string) bool {
	n := len(arr)
	for i := 0; i < n; i++ {
		if in == arr[i] {
			return true
		}
	}

	return false
}

// Contains determines whether an array includes a certain value among its entries
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

// NumFormat convert a number to string
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
