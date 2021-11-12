package opm

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type routeRegexp struct {
	regexp  *regexp.Regexp
	reverse string
	VarsN   []string
	VarsR   []*regexp.Regexp
}

func newRouteRegexp(path string) (*routeRegexp, error) {
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	idxs, err := braceIndices(path)
	if err != nil {
		return nil, err
	}

	defaultPattern := "[^/]+"

	varsN := make([]string, len(idxs)/2)
	varsR := make([]*regexp.Regexp, len(idxs)/2)

	pattern := bytes.NewBufferString("")
	pattern.WriteByte('^')
	reverse := bytes.NewBufferString("")

	var end int
	for i := 0; i < len(idxs); i += 2 {
		raw := path[end:idxs[i]]
		end = idxs[i+1]
		parts := strings.SplitN(path[idxs[i]+1:end-1], ":", 2)
		name := parts[0]
		patt := defaultPattern
		if len(parts) == 2 {
			patt = parts[1]
		}

		if name == "" || patt == "" {
			return nil, fmt.Errorf("missing name or pattern in %q", path[idxs[i]:end])
		}

		gn := "v" + strconv.Itoa(i/2)
		fmt.Fprintf(pattern, "%s(?P<%s>%s)", regexp.QuoteMeta(raw), gn, patt)
		fmt.Fprintf(reverse, "%s%%s", raw)

		varsN[i/2] = name
		varsR[i/2], err = regexp.Compile(fmt.Sprintf("^%s$", patt))
		if err != nil {
			return nil, err
		}
	}

	raw := path[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	pattern.WriteByte('$')

	reverse.WriteString(raw)

	reg, err := regexp.Compile(pattern.String())
	if err != nil {
		return nil, err
	}

	if reg.NumSubexp() != len(idxs)/2 {
		panic(fmt.Sprintf("route %s contains capture groups in its regexp.", path))
	}

	return &routeRegexp{
		regexp:  reg,
		reverse: reverse.String(),
		VarsN:   varsN,
		VarsR:   varsR,
	}, nil
}

func (r *routeRegexp) Math(req *http.Request) bool {
	var path = getPath(req)
	return r.regexp.MatchString(path)
}

func braceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int

	ls := len(s)
	for i := 0; i < ls; i++ {
		if s[i] == '{' {
			if level++; level == 1 {
				idx = i
			}
		}

		if s[i] == '}' {
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("unbalanced braces in %q", s)
			}
		}
	}

	if level != 0 {
		return nil, fmt.Errorf("unbalanced braces in %q", s)
	}

	return idxs, nil
}
