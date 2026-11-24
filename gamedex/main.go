package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
)

type Game struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Studio string   `json:"studio"`
	Genres []string `json:"genres"`
}

func main() {
	port := flag.String("port", "3000", "port du serveur")
	apiURL := flag.String("api", "http://localhost:8000", "URL de base API")
	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))


	// accueil
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<link rel="stylesheet" href="/static/css/style.css"><h1>Accueil GameDex</h1>`)
	})

	// appeller GET /api/games
	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(*apiURL + "/api/games")
		if err != nil {
			http.Error(w, "API inaccessible", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		var jeux []Game
		if err := json.NewDecoder(resp.Body).Decode(&jeux); err != nil {
			http.Error(w, "Erreur lecture JSON", http.StatusInternalServerError)
			return
		}

		for _, j := range jeux {
			fmt.Fprintf(w, "- [%d] %s (%d - %s)\n", j.ID, j.Name, j.Year, j.Studio)
		}
	})

	// appeller GET /api/games/{id}
	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if _, err := strconv.Atoi(idStr); err != nil {
			http.Error(w, "ID invalide", http.StatusBadRequest) // 400
			return
		}

		resp, err := http.Get(*apiURL + "/api/games/" + idStr)
		if err != nil {
			http.Error(w, "API inaccessible", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Si 404
		if resp.StatusCode == http.StatusNotFound {
			http.Error(w, "Jeu introuvable", http.StatusNotFound) // 404
			return
		}

		var jeu Game
		if err := json.NewDecoder(resp.Body).Decode(&jeu); err != nil {
			http.Error(w, "Erreur lecture JSON", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Jeu #%d : %s (%d) par %s\n", jeu.ID, jeu.Name, jeu.Year, jeu.Studio)
	})

	fmt.Println("Serveur sur http://localhost:" + *port)
	http.ListenAndServe(":"+*port, mux)
}
