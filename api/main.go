package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"uuid"

	"github.com/joho/godotenv"
)

type Note struct {
	Name string
	Text string
	Id   uuid.UUID
}

type Config struct {
	Port   string
	DbPort string
	Host   string
}

type Server struct {
	cfg    Config
	client *http.Client
	tmpl   *template.Template
}

func (s *Server) writeNote(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.tmpl.ExecuteTemplate(w, "new_note.html", nil)
	} else {
		name := r.FormValue("name")
		text := r.FormValue("text")
		log.Printf("form: name=%q text=%q", name, text)
		note := Note{name, text, uuid.Nil()}
		resp, err := s.dbPost(s.dbURL("/save"), note)
		if err != nil {
			http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
			return
		}
		log.Printf("%s", resp)
		http.Redirect(w, r, "/notes", http.StatusSeeOther)
	}
}

func (s *Server) listNotes(w http.ResponseWriter, r *http.Request) {
	body, err := s.dbGet(s.dbURL("/notes"))
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
	s.tmpl.ExecuteTemplate(w, "notes.html", notes)
}

func (s *Server) dbDel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("Id")

	req, err := http.NewRequest(http.MethodDelete, s.dbURL("/delete/"+id), nil)
	if err != nil {
		http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
		return
	}

	resp, err := s.makeRequest(req)
	if err != nil {
		http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
		return
	}
	log.Printf("%s", resp)
	http.Redirect(w, r, "/notes", http.StatusSeeOther)
}

func (s *Server) dbGet(endpoint string) ([]byte, error) {
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", "application/json")

	return s.makeRequest(request)
}

func (s *Server) dbPost(endpoint string, note Note) ([]byte, error) {
	data, err := json.Marshal(note)
	if err != nil {
		return nil, fmt.Errorf("dbPost: %w", err)
	}

	request, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("dbPost: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	return s.makeRequest(request)
}

func (s *Server) makeRequest(request *http.Request) ([]byte, error) {
	resp, err := s.client.Do(request)
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadConfig() Config {
	return Config{
		Port:   getEnv("GO_PORT", "8090"),
		DbPort: getEnv("DB_PORT", "8091"),
		Host:   getEnv("HOST", "localhost"),
	}
}

func NewServer(cfg Config) *Server {
	return &Server{
		cfg:    cfg,
		client: &http.Client{Timeout: 5 * time.Second},
		tmpl:   template.Must(template.ParseGlob("templates/*.html")),
	}
}

func (s *Server) dbURL(path string) string {
	return fmt.Sprintf("http://%s:%s%s", s.cfg.Host, s.cfg.DbPort, path)
}

func main() {
	godotenv.Load()

	cfg := loadConfig()
	srv := NewServer(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/new", srv.writeNote)
	mux.HandleFunc("/notes", srv.listNotes)
	mux.HandleFunc("/delete/{Id}", srv.dbDel)

	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
