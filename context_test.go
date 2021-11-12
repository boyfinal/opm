package opm

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkJSON(b *testing.B) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("userJSON"))
	w := httptest.NewRecorder()

	router := NewRouter()
	c := router.NewContext(w, r).(*context)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.JSON(http.StatusOK, strings.NewReader(`{"name": "Hanh"}`))
	}
}

func BenchmarkHTML(b *testing.B) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("userJSON"))
	w := httptest.NewRecorder()

	s := NewRouter()
	c := s.NewContext(w, r).(*context)

	b.ResetTimer()
	b.ReportAllocs()

	html := `
		<!DOCTYPE html>
		<html lang="en">
			<head>
				<meta charset="UTF-8" />
				<meta name="HandheldFriendly" content="true" />
				<meta name="viewport" content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=no" />
				<title>Test</title>
			</head>
			<body>
				<h1>Welcome</h1>
			</body>
		</html>
	`

	for i := 0; i < b.N; i++ {
		c.HTML(http.StatusOK, html)
	}
}

func BenchmarkFile(b *testing.B) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("userJSON"))
	w := httptest.NewRecorder()

	s := NewRouter()
	c := s.NewContext(w, r).(*context)

	b.ResetTimer()
	b.ReportAllocs()

	html := `
		<!DOCTYPE html>
		<html lang="en">
			<head>
				<meta charset="UTF-8" />
				<meta name="HandheldFriendly" content="true" />
				<meta name="viewport" content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=no" />
				<title>Test</title>
			</head>
			<body>
				<h1>Welcome</h1>
			</body>
		</html>
	`

	for i := 0; i < b.N; i++ {
		c.HTML(http.StatusOK, html)
	}
}
