package template

import (
	"html/template"

	"github.com/boyfinal/opm"
)

func rawhtml(str string) template.HTML {
	return template.HTML(str)
}

func format(v interface{}) string {
	return opm.NumFormat(v)
}

func add(a, b int) int {
	return a + b
}
