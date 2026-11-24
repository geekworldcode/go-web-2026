package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Game struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Studio string   `json:"studio"`
	Genres []string `json:"genres"`
}

type Tournament struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	GameID int    `json:"gameId"`
}

var httpClient = &http.Client{
	Timeout: 3 * time.Second,
}

func fetchJSON(url string, target any) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("API inaccessible ou trop lente (%w)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erreur HTTP %d sur %s", resp.StatusCode, url)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("JSON invalide reçu de %s (%w)", url, err)
	}
	return nil
}

func FetchGames(apiURL string) ([]Game, error) {
	var games []Game
	err := fetchJSON(apiURL+"/api/games", &games)
	return games, err
}

func FetchTournaments(apiURL string) ([]Tournament, error) {
	var tournaments []Tournament
	err := fetchJSON(apiURL+"/api/tournaments", &tournaments)
	return tournaments, err
}

func main() {
	port := flag.String("port", "3000", "port du serveur")
	apiURL := flag.String("api", "http://localhost:8000", "URL de base API")
	flag.Parse()

	// affichage dans le terminal
	games, err := FetchGames(*apiURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}
	tournaments, err := FetchTournaments(*apiURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}

	for _, g := range games {
		count := 0
		for _, t := range tournaments {
			if t.GameID == g.ID {
				count++
			}
		}
		fmt.Printf("- %s (%d) : %d tournoi(s)\n", g.Name, g.Year, count)
	}

	jeux := map[int]string{
		1: "The Legend of Zelda (1986) - Nintendo",
		2: "Hollow Knight (2017) - Team Cherry",
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// accueil
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<link rel="stylesheet" href="/static/css/style.css"><h1>Accueil GameDex</h1>`)
	})

	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		for id, nom := range jeux {
			fmt.Fprintf(w, "- [%d] %s\n", id, nom)
		}
	})

	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "ID invalide", http.StatusBadRequest) // 400
			return
		}
		jeu, ok := jeux[id]
		if !ok {
			http.Error(w, "Jeu introuvable", http.StatusNotFound) // 404
			return
		}
		fmt.Fprintf(w, "Jeu #%d : %s\n", id, jeu)
	})

	fmt.Println("Serveur sur http://localhost:" + *port)
	http.ListenAndServe(":"+*port, mux)
}