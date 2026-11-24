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

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// accueil
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<link rel="stylesheet" href="/static/css/style.css"><h1>Accueil GameDex</h1>`)
	})

	// appeller GET /api/games
	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		jeux, err := FetchGames(*apiURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
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
		var jeu Game
		err := fetchJSON(*apiURL+"/api/games/"+idStr, &jeu)
		// Si 404
		if err != nil {
			http.Error(w, "Jeu introuvable ou erreur API", http.StatusNotFound) // 404
			return
		}
		fmt.Fprintf(w, "Jeu #%d : %s (%d) par %s\n", jeu.ID, jeu.Name, jeu.Year, jeu.Studio)
	})

	fmt.Println("Serveur sur http://localhost:" + *port)
	http.ListenAndServe(":"+*port, mux)
}