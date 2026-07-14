package mocks

import (
	"time"

	"github.com/discruter/scratchpad/internal/models"
)

var mockPad = models.Pads{
	ID:      1,
	Title:   "An old silent pond",
	Content: "An old silent pond...",
	Created: time.Now(),
	Expires: time.Now(),
}

type PadsModel struct{}

func (m *PadsModel) Insert(title, content string, expires int) (int, error) {
	return 2, nil
}

func (m *PadsModel) Get(id int) (models.Pads, error) {
	switch id {
	case 1:
		return mockPad, nil
	default:
		return models.Pads{}, models.ErrNoRecord
	}
}

func (m *PadsModel) Latest() ([]models.Pads, error) {
	return []models.Pads{mockPad}, nil
}
