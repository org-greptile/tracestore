package app

import (
	_ "embed"
	"html/template"
	"net/http"
	"path"
	"strings"

	"example.com/acme/kit/kv/memberlist"
)

//go:embed memberlist_status.gohtml
var memberlistStatusPageHTML string

func memberlistStatusHandler(httpPathPrefix string, kvs *memberlist.KVInitService) http.Handler {
	templ := template.New("memberlist_status")
	templ.Funcs(map[string]interface{}{
		"AddPathPrefix": func(link string) string { return path.Join(httpPathPrefix, link) },
		"StringsJoin":   strings.Join,
	})
	template.Must(templ.Parse(memberlistStatusPageHTML))
	return memberlist.NewHTTPStatusHandler(kvs, templ)
}
