package template

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type Template struct {
	sync.Mutex

	dirbase   string
	dirview   string
	dirlayout string
	templates map[string]*template.Template
}

var mainTmpl = `{{ define "main" }}{{ template "base" . }}{{end}}`

func New(dir string) *Template {
	return &Template{
		dirbase:   dir,
		dirview:   "pages",
		dirlayout: "layout",
		templates: make(map[string]*template.Template),
	}
}

func (t *Template) Dirview(name string) *Template {
	t.dirview = name
	return t
}

func (t *Template) Dirlayout(name string) *Template {
	t.dirlayout = name
	return t
}

func (t *Template) Render(w io.Writer, name string, body interface{}) error {
	if err := t.Load(name); err != nil {
		return err
	}

	if tmpl, ok := t.templates[name]; ok {
		return tmpl.Execute(w, body)
	}

	return fmt.Errorf("the template %s does not exist", name)
}

func (t *Template) Load(name string) error {
	if t.templates[name] != nil {
		return nil
	}

	var files []string

	pattern := filepath.Clean(fmt.Sprintf("%s/%s/*.html", t.dirbase, t.dirlayout))
	layouts, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	files = append(files, layouts...)

	tmp := template.New("main")
	tmp.Funcs(template.FuncMap{
		"raw":    rawhtml,
		"format": format,
		"add":    add,
	})

	tmp, err = tmp.Parse(mainTmpl)
	if err != nil {
		return err
	}

	file := filepath.Clean(fmt.Sprintf("%s/%s/%s", t.dirbase, t.dirview, name))
	if _, err := os.Stat(file); err != nil {
		return err
	}

	files = append(files, file)
	t.templates[name], err = tmp.Clone()
	if err != nil {
		return err
	}

	t.Lock()
	t.templates[name] = template.Must(t.templates[name].ParseFiles(files...))
	t.Unlock()

	return nil
}

func (t *Template) Refesh() {
	t.templates = make(map[string]*template.Template)
}
