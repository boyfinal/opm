package opm

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type response struct {
	Message string `json:"message"`
}

const responseJSON = `{"message":"hello"}`

func BenchmarkJSON(b *testing.B) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router := NewRouter()
	c := router.NewContext(w, r).(*context)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.JSON(http.StatusOK, strings.NewReader(responseJSON))
	}
}

func BenchmarkHTML(b *testing.B) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
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
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	s := NewRouter()
	c := s.NewContext(w, r).(*context)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.File("_fixture/assets/demo.jpg")
	}
}

func assertContentTypeBody(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	if strings.Compare(rec.Header().Get(HeaderContentType), expected) != 0 {
		t.Fatalf("response content type should be %v, got: %q",
			expected, rec.Header().Get(HeaderContentType))
	}
}

func assertResponseBody(t *testing.T, rec *httptest.ResponseRecorder, expectedBody string) {
	if rec.Code != 200 {
		t.Fatalf("expected a status code of 200, got %v", rec.Code)
	}

	body, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("unexpected error reading body: %v", err)
	}

	if !bytes.Equal(body, []byte(expectedBody)) {
		t.Fatalf("response should be %s, was: %q", expectedBody, string(body))
	}
}

func assertError(t *testing.T, err error) bool {
	if err == nil {
		t.Fatalf("An error is expected but got nil.")
		return false
	}

	return true
}

func assertNoError(t *testing.T, err error) bool {
	if err != nil {
		t.Fatalf("Received unexpected error:\n%+v", err)
		return false
	}

	return true
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func TestContext(t *testing.T) {
	o := NewRouter()
	req := httptest.NewRequest("POST", "/", strings.NewReader("userJSON"))
	rec := httptest.NewRecorder()
	c := o.NewContext(rec, req)

	if c.Request() == nil {
		t.Fatal("request is null")
	}

	if c.Response() == nil {
		t.Fatal("response is null")
	}

	tmpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.name}}!")),
	}

	c.SetRenderer(tmpl)
	c.Set("name", "Saitama")
	err := c.Render(http.StatusOK, "hello")
	if assertNoError(t, err) {
		assertContentTypeBody(t, rec, MIMETextHTMLCharsetUTF8)
		assertResponseBody(t, rec, "Hello, Saitama!")
	}

	// JSON
	rec = httptest.NewRecorder()
	c = o.NewContext(rec, req).(*context)
	err = c.JSON(http.StatusOK, response{"hello"})
	if assertNoError(t, err) {
		assertContentTypeBody(t, rec, MIMEApplicationJSONCharsetUTF8)
		assertResponseBody(t, rec, responseJSON)
	}

	// JSON (error)
	rec = httptest.NewRecorder()
	c = o.NewContext(rec, req).(*context)
	err = c.JSON(http.StatusOK, make(chan bool))
	assertError(t, err)

	// String
	rec = httptest.NewRecorder()
	c = o.NewContext(rec, req).(*context)
	err = c.String(http.StatusOK, "hello")
	if assertNoError(t, err) {
		assertContentTypeBody(t, rec, MIMETextPlainCharsetUTF8)
		assertResponseBody(t, rec, "hello")
	}

	// NoContent
	rec = httptest.NewRecorder()
	c = o.NewContext(rec, req).(*context)
	err = c.NoContent(http.StatusOK)
	if assertNoError(t, err) {
		assertResponseBody(t, rec, "")
	}

	// NoContent
	rec = httptest.NewRecorder()
	c = o.NewContext(rec, req).(*context)
	err = c.NoContent(http.StatusOK)
	if assertNoError(t, err) {
		assertResponseBody(t, rec, "")
	}

	// HTML
	html := `<!DOCTYPE html><html lang="en"><head><title>OnePuchMan</title></head><body></body></html>`
	rec = httptest.NewRecorder()
	c = o.NewContext(rec, req).(*context)
	err = c.HTML(http.StatusOK, html)
	if assertNoError(t, err) {
		assertContentTypeBody(t, rec, MIMETextHTMLCharsetUTF8)
		assertResponseBody(t, rec, html)
	}

	// Stream
	rec = httptest.NewRecorder()
	c = o.NewContext(rec, req).(*context)
	r := strings.NewReader("response from a stream")
	err = c.Stream(http.StatusOK, "application/octet-stream", r)
	if assertNoError(t, err) {
		assertContentTypeBody(t, rec, "application/octet-stream")
		assertResponseBody(t, rec, "response from a stream")
	}

	// File
	rec = httptest.NewRecorder()
	c = o.NewContext(rec, req).(*context)
	err = c.File("_fixture/assets/demo.jpg")
	if assertNoError(t, err) {
		assertContentTypeBody(t, rec, "image/jpeg")
		if rec.Body.Len() != 98709 {
			t.Fatalf("file len %v, response %v", 98709, rec.Body.Len())
		}
	}
}

func BenchmarkNoContent(b *testing.B) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router := NewRouter()
	c := router.NewContext(w, r).(*context)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.NoContent(http.StatusOK)
	}
}

func BenchmarkBlob(b *testing.B) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router := NewRouter()
	c := router.NewContext(w, r).(*context)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Blob(http.StatusOK, MIMETextPlainCharsetUTF8, []byte("opm"))
	}
}

func BenchmarkString(b *testing.B) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router := NewRouter()
	c := router.NewContext(w, r).(*context)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.String(http.StatusOK, "opm")
	}
}
