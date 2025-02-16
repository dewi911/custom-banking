package service

import (
	"context"
	"custom-banking/internal/models"
	"github.com/pkg/errors"
)

type Event struct {
	eventRepository EventRepository
}

func NewEvent(eventRepository EventRepository) *Event {
	return &Event{eventRepository}
}

func (e *Event) GetEventList(ctx context.Context, userID int) ([]models.Event, error) {
	lsit, err := e.eventRepository.GetEventsList(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "GetEventList")
	}

	return lsit, nil
}
