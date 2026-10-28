package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
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

func listNotes(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("notes.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request, err := http.NewRequest("GET", "http://localhost:8091/notes", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	request.Header.Add("Accept", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "upstream error: "+resp.Status, http.StatusBadGateway)
		fmt.Println(io.ReadAll(resp.Body))
		return
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	var notes []Note
	err = json.Unmarshal(body, &notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, notes)
}

func main() {
	http.HandleFunc("/new", writeNote)
	http.HandleFunc("/notes", listNotes)
	http.ListenAndServe(":8090", nil)
}
