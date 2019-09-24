package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/go-chi/chi"

	cas "github.com/bangwork-utils/cas.v2"
)

type templateBinding struct {
	Username   string
	Attributes cas.UserAttributes
}

var helpInfo = `
Usage:
	%s [listen_addr] [cas-server-url]
Example:
	%s :9001 https://casdemo.myones.net:8443/cas
`

func main() {
	if len(os.Args) != 3 {
		fmt.Printf(helpInfo, os.Args[0], os.Args[0])
		return
	}

	casURL := os.Args[2]
	addr := os.Args[1]

	url, _ := url.Parse(casURL)
	client := cas.NewClient(&cas.Options{URL: url})

	root := chi.NewRouter()
	root.Use(client.Handler)

	//这句为新增代码
	server := &http.Server{
		Addr:    addr,
		Handler: client.Handle(root),
	}

	root.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")

		tmpl, err := template.New("index.html").Parse(index_html)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, error_500, err)
			return
		}

		binding := &templateBinding{
			Username:   cas.Username(r),
			Attributes: cas.Attributes(r),
		}

		log.Printf("attributes: %s %v", binding.Username, binding.Attributes)

		html := new(bytes.Buffer)
		if err := tmpl.Execute(html, binding); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, error_500, err)
			return
		}

		html.WriteTo(w)
	})

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)

	}
}

const index_html = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8"> 
    <title>CAS 返回参数测试</title>
  </head>
  <body>
	<h1>CAS 返回参数测试 <a href="/logout">Logout</a></h1>
    <ul>{{range $key, $values := .Attributes}}
      <li>{{$len := len $values}}{{$key}}:{{if gt $len 1}}
        <ul>{{range $values}}
          <li>{{.}}</li>{{end}}
        </ul>
      {{else}} {{index $values 0}}{{end}}</li>{{end}}
    </ul>
  </body>
</html>
`

const error_500 = `<!DOCTYPE html>
<html>
  <head>
    <title>Error 500</title>
  </head>
  <body>
    <h1>Error 500</h1>
    <p>%v</p>
  </body>
</html>
`
