package tst_logic

import (
	"avito_intern_dev/models"
	"math/rand"
)

func GenerateUsers(candidates []models.User) []models.User {
	if len(candidates) < 2 {
		return candidates
	} else {
		rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
		candidates = candidates[:2]
		return candidates

	}
}
