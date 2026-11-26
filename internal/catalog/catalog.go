package catalog

import (
	"sort"

	"go-web-2026/internal/api"
)

func Genres(games []api.Game) []string {
	seen := make(map[string]bool)
	var genres []string
	for _, game := range games {
		for _, genre := range game.Genres {
			if !seen[genre] {
				seen[genre] = true
				genres = append(genres, genre)
			}
		}
	}
	sort.Strings(genres)
	return genres
}

func TournamentsForGame(tournaments []api.Tournament, gameID int) []api.Tournament {
	var result []api.Tournament
	for _, t := range tournaments {
		if t.GameID == gameID {
			result = append(result, t)
		}
	}
	return result
}
