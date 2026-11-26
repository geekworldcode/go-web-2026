package server

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"go-web-2026/internal/api"
	"go-web-2026/internal/catalog"
	"go-web-2026/internal/store"
	"go-web-2026/web"
)

type JeuxPage struct {
	Games  []api.Game
	Query  string
	Genre  string
	Genres []string
}

type JeuPage struct {
	Game       api.Game
	Tournament []api.Tournament
	Likes      int
}

type ErrorPage struct {
	Code    int
	Message string
}

type Server struct {
	apiURL string
	mux    *http.ServeMux
	tmpl   *template.Template
}

func RunServer() {
	port := flag.String("port", "3000", "port du serveur")
	apiURL := flag.String("api", "http://localhost:8081", "URL de base API")
	flag.Parse()

	srv := &Server{
		apiURL: *apiURL,
		mux:    http.NewServeMux(),
		tmpl:   template.Must(template.ParseFS(web.Assets, "templates/*.html")),
	}

	srv.mux.Handle("GET /static/", http.FileServerFS(web.Assets))
	srv.Home()
	srv.Games()
	srv.GameByID()
	srv.LikeGame()

	fmt.Println("Serveur sur http://localhost:" + *port)
	if err := http.ListenAndServe(":"+*port, srv.mux); err != nil {
		fmt.Println("Erreur serveur :", err)
	}
}

func (s *Server) renderError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	if err := s.tmpl.ExecuteTemplate(w, "erreur.html", ErrorPage{
		Code:    code,
		Message: message,
	}); err != nil {
		http.Error(w, message, code)
	}
}

func (s *Server) Home() {
	s.mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<link rel="stylesheet" href="/static/css/style.css"><h1>Accueil GameDex</h1><a href="/jeux">Voir les jeux</a>`)
	})
}

func (s *Server) Games() {
	s.mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		query := strings.TrimSpace(params.Get("q"))
		genre := strings.TrimSpace(params.Get("genre"))

		jeux, err := api.FetchGames(s.apiURL, "", "")
		if err != nil {
			s.renderError(w, http.StatusBadGateway, "API indisponible : "+err.Error())
			return
		}

		data := JeuxPage{
			Games:  jeux,
			Query:  query,
			Genre:  genre,
			Genres: catalog.Genres(jeux),
		}
		if query != "" || genre != "" {
			data.Games, err = api.FetchGames(s.apiURL, query, genre)
			if err != nil {
				s.renderError(w, http.StatusBadGateway, "API indisponible : "+err.Error())
				return
			}
		}

		if err := s.tmpl.ExecuteTemplate(w, "jeux.html", data); err != nil {
			http.Error(w, "Erreur de rendu template : "+err.Error(), http.StatusInternalServerError)
		}
	})
}

func (s *Server) GameByID() {
	s.mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.renderError(w, http.StatusBadRequest, "ID invalide")
			return
		}

		jeu, err := api.FetchGameByID(s.apiURL, id)
		if err != nil {
			s.renderError(w, http.StatusNotFound, "Jeu introuvable ou erreur API")
			return
		}

		allTournaments, err := api.FetchTournaments(s.apiURL)
		if err != nil {
			s.renderError(w, http.StatusBadGateway, "API indisponible pour les tournois")
			return
		}

		data := JeuPage{
			Game:       jeu,
			Tournament: catalog.TournamentsForGame(allTournaments, id),
			Likes:      store.Likes(id),
		}

		if err := s.tmpl.ExecuteTemplate(w, "jeu.html", data); err != nil {
			http.Error(w, "Erreur de rendu template : "+err.Error(), http.StatusInternalServerError)
		}
	})
}

func (s *Server) LikeGame() {
	s.mux.HandleFunc("POST /jeux/{id}/like", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.renderError(w, http.StatusBadRequest, "ID invalide")
			return
		}
		if _, err := api.FetchGameByID(s.apiURL, id); err != nil {
			s.renderError(w, http.StatusNotFound, "Jeu introuvable ou erreur API")
			return
		}

		store.AddLike(id)
		http.Redirect(w, r, "/jeux/"+strconv.Itoa(id), http.StatusSeeOther)
	})
}
