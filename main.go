package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "embed"

	_ "github.com/lib/pq"
)

//go:embed template.html
var templatePage string
var page *template.Template
var title string = "Hello World"

type TemplateData struct {
	Title        string
	IsStandalone bool
	Hostname     string
	Messages     []string
}

func createStandalone() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Execute(w, TemplateData{
			Title:        title,
			IsStandalone: true,
			Hostname:     getHostname(),
		})
	})
}

func createDatabase(dsn string) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS message (data text)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		message := r.FormValue("message")
		_, err := db.Exec("INSERT INTO message (data) VALUES ($1)", message)
		if err != nil {
			log.Print(err)
			w.Write([]byte("Error: " + err.Error()))
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		_, err = db.Exec("DELETE FROM message")
		if err != nil {
			log.Print(err)
			w.Write([]byte("Error: " + err.Error()))
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("select data from message")
		if err != nil {
			log.Print(err)
			w.Write([]byte("Error: " + err.Error()))
			return
		}
		defer rows.Close()

		var messages []string
		for rows.Next() {
			var message string
			err = rows.Scan(&message)
			if err != nil {
				panic(err)
			}
			messages = append(messages, message)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Execute(w, TemplateData{
			Title:        title,
			IsStandalone: false,
			Hostname:     getHostname(),
			Messages:     messages,
		})
	})
}

func main() {
	addr := flag.String("addr", ":8080", "server address")
	titleFile := flag.String("title-file", "", "title file")
	flag.Parse()

	if *titleFile != "" {
		data, err := os.ReadFile(*titleFile)
		if err != nil {
			log.Fatal(err)
		}
		title = string(data)
	}

	page = template.Must(template.New("page").Parse(templatePage))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		createStandalone()
	} else {
		createDatabase(dsn)
	}

	log.Print("Listening on " + *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Print(err)
		return "unknown"
	}
	return hostname
}
