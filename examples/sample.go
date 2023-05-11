package main

import (
	"html/template"
	"net/http"
	"time"

	"github.com/go-faker/faker/v4"
	livelygo "github.com/lukasmwerner/lively.go"
)

func main() {

	http.HandleFunc("/lively.js", livelygo.Javascript)

	tpl := template.Must(template.New("sample").Parse(`<h1 lively-bind="title"></h1>
	<div lively-bind="colorful">
		<h1>hello there! </h1>	
		<p> inital render </p>
	</div>
	<script src="/lively.js" type="module"></script>`))
	http.Handle("/sample-page", livelygo.NewPage(func(w http.ResponseWriter, r *http.Request) {
		go func() {
			p := livelygo.WaitForPage(r)
			p.SetVar("title", `<h1 lively-bind="title">Hello, World!</h1>`)
			go func() {
				ch := time.NewTicker(3 * time.Second)
				for {
					<-ch.C
					p.SetVar("colorful", `<div lively-bind="colorful"><h2 style="color:blue;"> Hello, `+faker.FirstName()+" "+faker.LastName()+`</h2></div>`)
				}
			}()
			ch := time.NewTicker(time.Second)
			for {
				<-ch.C
				p.SetVar("title", `<h1 lively-bind="title">Hello, World! `+time.Now().Format("03:04:05")+`</h1>`)
			}
		}()
		tpl.Execute(w, nil)
	}))

	http.ListenAndServe(":8080", nil)
}
