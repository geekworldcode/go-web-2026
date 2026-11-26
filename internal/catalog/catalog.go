package catalog

import (
	"strings"

	"go-web-2026/internal/api"
)

func FilterGames(games []api.Game, query string) []api.Game {
	if query == "" {
		return games
	}

	qLower := strings.ToLower(query)
	var filtered []api.Game
	for _, g := range games {
		if strings.Contains(strings.ToLower(g.Name), qLower) {
			filtered = append(filtered, g)
		}
	}
	return filtered
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