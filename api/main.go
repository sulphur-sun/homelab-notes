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
	"uuid"
)

type Note struct {
	Name string
	Text string
	Id   uuid.UUID
}

var client = &http.Client{
	Timeout: 5 * time.Second,
}
var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func writeNote(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "new_note.html", nil)
	} else {
		name := r.FormValue("name")
		text := r.FormValue("text")
		log.Printf("form: name=%q text=%q", name, text)
		note := Note{name, text, uuid.Nil()}
		resp, err := dbPost("http://localhost:8091/save", note)
		if err != nil {
			http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
			return
		}
		log.Printf("%s", resp)
		http.Redirect(w, r, "/notes", http.StatusSeeOther)
	}
}

func listNotes(w http.ResponseWriter, r *http.Request) {
	body, err := dbGet("http://localhost:8091/notes")
	if err != nil {
		http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
		return
	}

	var notes []Note
	err = json.Unmarshal(body, &notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "notes.html", notes)
}

func dbGet(endpoint string) ([]byte, error) {
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", "application/json")

	return makeRequest(request)
}

func dbPost(endpoint string, note Note) ([]byte, error) {
	data, err := json.Marshal(note)
	if err != nil {
		return nil, fmt.Errorf("dbPost: %w", err)
	}

	request, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("dbPost: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	return makeRequest(request)
}

func makeRequest(request *http.Request) ([]byte, error) {
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("makeRequest: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("makeRequest: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("upstream: %s", body)
		return nil, fmt.Errorf("status: %s", resp.Status)
	}
	return body, nil
}

func main() {
	http.HandleFunc("/new", writeNote)
	http.HandleFunc("/notes", listNotes)
	log.Fatal(http.ListenAndServe(":8090", nil))
}
