package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

type Note struct {
	Name string
	Text string
}

func writeNote(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("new_note.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			println(err)
		}
		tmpl.Execute(w, nil)
	} else {
		name := r.FormValue("name")
		text := r.FormValue("text")
		log.Printf("form: name=%q text=%q", name, text)
		note := Note{name, text}
		data, err := json.Marshal(note)
		log.Printf("marshaled: %s (err=%v)", string(data), err)
		if err != nil {
			println(err)
		} else {
			request, err := http.NewRequest("POST", "http://localhost:8091/save", bytes.NewReader(data))
			request.Header.Set("Content-Type", "application/json")
			if err != nil {
				println(err)
			} else {
				http.DefaultClient.Do(request)
			}
		}
	}
}

func main() {
	http.HandleFunc("/new", writeNote)
	http.ListenAndServe(":8090", nil)
}
