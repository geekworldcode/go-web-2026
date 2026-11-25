package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	City   string `json:"city"`
	Date   string `json:"date"`
}

type JeuxPage struct {
	Games []Game
	Query string
}

type JeuPage struct {
	Game       Game
	Tournament []Tournament
}

type ErrorPage struct {
	Code    int
	Message string
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

	// Affichage dans le terminal
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

	// Templates
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	// Gestion des erreurs HTTP
	renderError := func(w http.ResponseWriter, code int, message string) {
		w.WriteHeader(code)
		if err := tmpl.ExecuteTemplate(w, "erreur.html", ErrorPage{
			Code:    code,
			Message: message,
		}); err != nil {
			http.Error(w, message, code)
		}
	}

	// Accueil
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<link rel="stylesheet" href="/static/css/style.css"><h1>Accueil GameDex</h1><a href="/jeux">Voir les jeux</a>`)
	})

	// Appeler GET /api/games
	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		jeux, err := FetchGames(*apiURL)
		if err != nil {
			renderError(w, http.StatusBadGateway, "API indisponible : "+err.Error())
			return
		}

		query := strings.TrimSpace(r.URL.Query().Get("q"))
		var filtered []Game
		if query != "" {
			qLower := strings.ToLower(query)
			for _, g := range jeux {
				if strings.Contains(strings.ToLower(g.Name), qLower) {
					filtered = append(filtered, g)
				}
			}
		} else {
			filtered = jeux
		}

		data := JeuxPage{
			Games: filtered,
			Query: query,
		}

		if err := tmpl.ExecuteTemplate(w, "jeux.html", data); err != nil {
			http.Error(w, "Erreur de rendu template : "+err.Error(), http.StatusInternalServerError)
		}
	})

	// Appeler GET /api/games/{id}
	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			renderError(w, http.StatusBadRequest, "ID invalide")
			return
		}

		var jeu Game
		if err := fetchJSON(*apiURL+"/api/games/"+idStr, &jeu); err != nil {
			renderError(w, http.StatusNotFound, "Jeu introuvable ou erreur API")
			return
		}

		allTournaments, err := FetchTournaments(*apiURL)
		if err != nil {
			renderError(w, http.StatusBadGateway, "API indisponible pour les tournois")
			return
		}

		var gameTournaments []Tournament
		for _, t := range allTournaments {
			if t.GameID == id {
				gameTournaments = append(gameTournaments, t)
			}
		}

		data := JeuPage{
			Game:       jeu,
			Tournament: gameTournaments,
		}

		if err := tmpl.ExecuteTemplate(w, "jeu.html", data); err != nil {
			http.Error(w, "Erreur de rendu template : "+err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Println("Serveur sur http://localhost:" + *port)
	http.ListenAndServe(":"+*port, mux)
}
//