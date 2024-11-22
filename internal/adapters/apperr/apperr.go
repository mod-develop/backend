package apperr

import "errors"

var (
	ErrDataNotFound = errors.New("data not found")

	// player quest
	ErrPlayerQuestStatusExists = errors.New("quest player status already exists")
)
