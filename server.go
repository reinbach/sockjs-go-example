package main

import (
	"github.com/igm/sockjs-go/sockjs"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var folder_static = "./"
var folder_templates = "./templates/"

var data interface{}

func main() {
	log.Println("server started...")

	sockjs.Install("/sockjs/echo", sockEchoHandler, sockjs.DefaultConfig)
	sockjs.Install("/sockjs/ping", sockEchoHandler, sockjs.DefaultConfig)

	http.Handle("/static/", http.FileServer(http.Dir(folder_static)))
	http.HandleFunc("/", pageHandler)

	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// handle favicon, by ignoring
	if path == "/favicon.ico" {
		return
	}

	file := folder_templates + "index.html"
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	if path != "" {
		file = folder_templates + path[1:] + ".html"
	}

	t, err := template.ParseFiles(folder_templates+"base.html", file)
	if err != nil {
		log.Println("template parse error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func sockEchoHandler(conn sockjs.Conn) {
	log.Println("echo session")
	for {
		if msg, err := conn.ReadMessage(); err != nil {
			log.Println("getting err:", err)
			return
		} else {
			go func() { conn.WriteMessage(msg) }()
		}
	}
}
