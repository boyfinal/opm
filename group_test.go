package opm

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	g := Make().Group("/group")
	h := func(Context) error { return nil }
	g.CONNECT("/", h)
	g.DELETE("/", h)
	g.GET("/", h)
	g.HEAD("/", h)
	g.OPTIONS("/", h)
	g.PATCH("/", h)
	g.POST("/", h)
	g.PUT("/", h)
	g.TRACE("/", h)
	g.Any("/", h)
	g.File("/demo", "_fixture/assets/demo.jpg")
}

func TestGroupFile(t *testing.T) {
	r := Make()
	g := r.Group("/group")

	g.File("/demo", "_fixture/assets/demo.jpg")

	expectedData, err := ioutil.ReadFile("_fixture/assets/demo.jpg")
	assert.Nil(t, err)

	req := httptest.NewRequest(http.MethodGet, "/group/demo", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, expectedData, rec.Body.Bytes())
}
