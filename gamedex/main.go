package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
)

func main() {
	port := flag.String("port", "3000", "port du serveur")
	flag.Parse()

	jeux := map[int]string{
		1: "The Legend of Zelda - Nintendo",
		2: "Hollow Knight - Team Cherry",
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// accueil html et css basique
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<link rel="stylesheet" href="/static/css/style.css"><h1>Accueil GameDex</h1>`)
	})

	// afficher les jeux
	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		for id, nom := range jeux {
			fmt.Fprintf(w, "- [%d] %s\n", id, nom)
		}
	})

	// details d'un jeu 
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