package utils

import (
	"aegis/poc/models"
)

func FindClientIndex(id string, queue []*models.Client) (int, bool) {
	for i, c := range queue {
		if c.ID == id {
			return i, true
		}
	}
	return -1, false
}
