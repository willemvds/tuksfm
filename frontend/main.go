package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	textTemplate "text/template"

	"golang.org/x/net/websocket"

	"gopkg.in/fsnotify.v1"
)

var templateNames []string = []string{
	"index.html",
}
var templates map[string]*template.Template

func GetTemplate(name string) *template.Template {
	return templates[name]
}

func LoadTemplate(name string) error {
	tmpl, err := template.New(name).ParseFiles(name)
	if err != nil {
		log.Println("failed to load template:", err)
		return err
	}
	templates[name] = tmpl
	return nil
}

func IndexHandler(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	tt, err := textTemplate.New("test").Parse("The simple life of rock and roll is like {{.}}\n")
	fmt.Println(tt)
	if err != nil {
		log.Println(err)
	}
	err = tt.Execute(os.Stdout, "a nice niceness")
	if err != nil {
		log.Println("text template failure", err)
	}
	//err = indexTmpl.ExecuteTemplate(res, "ROOT", map[string]interface{}{
	tmpl := GetTemplate("index.html")
	if tmpl == nil {
		http.Error(res, "Internal Server ERRORDAMN!", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(res, "ROOT", map[string]interface{}{
		"greeting": "Tuks Playlist",
		"last10": []string{
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
			"Red Jumpsuit Apparatus - Guardian Angel",
		},
	})
	if err != nil {
		log.Println("html template failure", err)
	}
}

func EchoServer(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

func main() {
	templates = make(map[string]*template.Template)
	for i := range templateNames {
		name := templateNames[i]
		LoadTemplate(name)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		print("create watcher error:", err)
		return
	}
	defer watcher.Close()

	for name := range templates {
		err = watcher.Add(name)
		if err != nil {
			log.Println("add watch error:", err)
		}
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Write file event:", event.Name)
					err = LoadTemplate(event.Name)
					if err != nil {
						log.Printf("Load/Reload of template (%s) failed: %s\n", event.Name, err)
					}
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("Rename file event:", event.Name)
					err = LoadTemplate(event.Name)
					if err != nil {
						log.Printf("Load/Reload of template (%s) failed: %s\n", event.Name, err)
					}
				}
			case err := <-watcher.Errors:
				log.Println("event error:", err)
			}
		}
	}()

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/favicon.ico", http.NotFound)

	http.Handle("/ws", websocket.Handler(EchoServer))

	err = http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Println(err)
	}
}
