package api

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	City   string `json:"city"`
	Date   string `json:"date"`
}

var httpClient = &http.Client{
	Timeout: 3 * time.Second,
}

func FetchJSON(url string, target any) error {
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
	err := FetchJSON(apiURL+"/api/games", &games)
	return games, err
}

func FetchGameByID(apiURL string, id int) (Game, error) {
	var game Game
	err := FetchJSON(apiURL+"/api/games/"+strconv.Itoa(id), &game)
	return game, err
}

func FetchTournaments(apiURL string) ([]Tournament, error) {
	var tournaments []Tournament
	err := FetchJSON(apiURL+"/api/tournaments", &tournaments)
	return tournaments, err
}