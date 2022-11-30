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

	FuncMap   template.FuncMap
	dirBase   string
	dirView   string
	dirLayout string
	templates map[string]*template.Template
}

var mainTmpl = `{{ define "main" }}{{ template "base" . }}{{end}}`

func New(dir string) *Template {
	return &Template{
		dirBase:   dir,
		dirView:   "pages",
		dirLayout: "layout",
		templates: make(map[string]*template.Template),
		FuncMap:   template.FuncMap{},
	}
}

func (t *Template) DirView(name string) *Template {
	t.dirView = name
	return t
}

func (t *Template) DirLayout(name string) *Template {
	t.dirLayout = name
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

func (t *Template) SetFuncMap(fnMap template.FuncMap) {
	t.FuncMap = fnMap
}

func (t *Template) Load(name string) error {
	if t.templates[name] != nil {
		return nil
	}

	var files []string

	pattern := filepath.Clean(fmt.Sprintf("%s/%s/*.html", t.dirBase, t.dirLayout))
	layouts, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	files = append(files, layouts...)

	tmp := template.New("main")

	// Add template func map basics
	tmp = tmp.Funcs(template.FuncMap{
		"raw":    rawhtml,
		"format": format,
		"add":    add,
	})

	tmp = tmp.Funcs(t.FuncMap)

	tmp, err = tmp.Parse(mainTmpl)
	if err != nil {
		return err
	}

	file := filepath.Clean(fmt.Sprintf("%s/%s/%s", t.dirBase, t.dirView, name))
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

func (t *Template) Refresh() {
	t.templates = make(map[string]*template.Template)
}
