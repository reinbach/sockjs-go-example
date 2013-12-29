package main

import (
	"fmt"
	"github.com/igm/sockjs-go/sockjs"
	"html/template"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

var folder_static = "./"
var folder_templates = "./templates/"

var data interface{}

func main() {
	log.Println("server started...")

	sockjs.Install("/sockjs/echo", sockEchoHandler, sockjs.DefaultConfig)
	sockjs.Install("/sockjs/ping", sockPingHandler, sockjs.DefaultConfig)
	sockjs.Install("/sockjs/startstop", sockStartStopHandler, sockjs.DefaultConfig)
	sockjs.Install("/sockjs/sine", sockSineHandler, sockjs.DefaultConfig)

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

func sockPingHandler(conn sockjs.Conn) {
	log.Println("ping session")
	for {
		if msg, err := conn.ReadMessage(); err != nil {
			log.Println("getting err:", err)
			return
		} else {
			if string(msg) == `"ping"` {
				go func() {
					conn.WriteMessage([]byte(`"pong"`))
				}()
			}
		}
	}
}

func sockStartStopHandler(conn sockjs.Conn) {
	log.Println("start/stop session")

	ticker := time.NewTicker(time.Second)
	open := true

	for {
		if msg, err := conn.ReadMessage(); err != nil {
			log.Println("getting err: ", err)
			ticker.Stop()
			open = false
			return
		} else {
			if string(msg) == `"start"` {
				if !open {
					ticker = time.NewTicker(time.Second)
					open = true
				}
				go func() {
					for t := range ticker.C {
						conn.WriteMessage([]byte(fmt.Sprintf(`"%d"`, t.Second())))
					}
				}()
			} else if string(msg) == `"stop"` {
				ticker.Stop()
				open = false
			}
		}
	}
}

func sockSineHandler(conn sockjs.Conn) {
	log.Println("sine session")

	var x, y float64
	var sine string

	ticker := time.NewTicker(time.Second)
	open := true

	for {
		if msg, err := conn.ReadMessage(); err != nil {
			log.Println("getting err:", err)
			ticker.Stop()
			open = false
			return
		} else {
			if string(msg) == `"start"` {
				if !open {
					ticker = time.NewTicker(time.Second)
					open = true
				}
				go func() {
					for t := range ticker.C {
						x = float64(t.Nanosecond()) / 1000
						y = 2.5 * (1 + math.Sin(x))
						sine = fmt.Sprintf(`{"x": "%f", "y": "%f"}`, x, y)
						log.Printf(sine)
						conn.WriteMessage([]byte(sine))
					}
				}()
			} else if string(msg) == `"stop"` {
				log.Println("stop sine")
				ticker.Stop()
				open = false
			}
		}
	}
}
