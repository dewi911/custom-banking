package repository

import (
	"context"
	"custom-banking/internal/models"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Event struct {
	db *sqlx.DB
}

func NewEvent(db *sqlx.DB) *Event {
	return &Event{db}
}

func (e *Event) CreateEvent(ctx context.Context, event models.Event) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "event",
		"method":     "CreateEvent",
		"event":      event,
	}

	mEvent := models.Event{
		UserID:   event.UserID,
		Type:     event.Type,
		Message:  event.Message,
		Metadata: event.Metadata,
	}

	query := "INSERT INTO event (user_id, type, metadata, time) VALUES ($1, $2, $3, now())"

	if _, err := e.db.ExecContext(ctx, query, mEvent.UserID, mEvent.Type, mEvent.Metadata); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("Failed to create event")

		return errors.Wrap(err, fmt.Sprintf("Failed to create event with user id %s", mEvent.UserID))
	}

	return nil
}

func (e *Event) GetEventsList(ctx context.Context, userID int) ([]models.Event, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "event",
		"method":     "GetEventsList",
		"userId":     userID,
	}

	query := "SELECT * FROM event WHERE user_id = $1 ORDER BY time DESC"

	rows, err := e.db.QueryContext(ctx, query, userID)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting events list query error")

		return nil, errors.Wrap(err, fmt.Sprintf("execution getting events list query error"))
	}

	var eventsList []models.Event
	for rows.Next() {
		var mevent models.Event
		if err = rows.Scan(&mevent); err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("scanning event row error")

			return nil, errors.Wrap(err, "scanning event row error")
		}

		eventType, err := models.NewEventTypeFromString(mevent.Type.String())
		if err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("scanning event row error")

			return nil, errors.Wrap(err, "scanning event row error")
		}

		eventsList = append(eventsList, models.Event{
			ID:       mevent.ID,
			UserID:   mevent.UserID,
			Type:     eventType,
			Message:  mevent.Message,
			Metadata: mevent.Metadata,
			DateTime: mevent.DateTime,
		})
	}

	return eventsList, nil
}
