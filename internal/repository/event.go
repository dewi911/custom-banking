package repository

import (
	"context"
	"custom-banking/internal/models"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
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

	var jsonMetadata []byte
	var err error

	if mEvent.Metadata != nil {
		jsonMetadata, err = json.Marshal(mEvent.Metadata)
		if err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("Failed to marshal metadata to JSON")
			return errors.Wrap(err, "Failed to marshal metadata to JSON")
		}
	}

	if _, err := e.db.ExecContext(ctx, query, mEvent.UserID, mEvent.Type, jsonMetadata); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("Failed to create event")

		return errors.Wrap(err, fmt.Sprintf("Failed to create event with user id %d", mEvent.UserID))
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

	rows, err := e.db.QueryxContext(ctx, query, userID)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting events list query error")

		return nil, errors.Wrap(err, fmt.Sprintf("execution getting events list query error"))
	}
	defer rows.Close()

	var eventsList []models.Event
	for rows.Next() {
		var mevent struct {
			ID       int             `db:"id"`
			UserID   int             `db:"user_id"`
			Type     string          `db:"type"`
			Message  string          `db:"message"`
			Metadata json.RawMessage `db:"metadata"`
			DateTime time.Time       `db:"time"`
		}

		if err = rows.StructScan(&mevent); err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("scanning event row error")

			return nil, errors.Wrap(err, "scanning event row error")
		}

		eventType, err := models.NewEventTypeFromString(mevent.Type)
		if err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("scanning event row error")

			return nil, errors.Wrap(err, "scanning event row error")
		}
		var metadata map[string]interface{}
		if len(mevent.Metadata) > 0 {
			if err := json.Unmarshal(mevent.Metadata, &metadata); err != nil {
				logrus.WithError(err).
					WithFields(fields).
					Error("unmarshaling metadata error")

				return nil, errors.Wrap(err, "unmarshaling metadata error")
			}
		}

		eventsList = append(eventsList, models.Event{
			ID:       mevent.ID,
			UserID:   mevent.UserID,
			Type:     eventType,
			Message:  mevent.Message,
			Metadata: metadata,
			DateTime: mevent.DateTime,
		})
	}

	if err = rows.Err(); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("error iterating over result rows")

		return nil, errors.Wrap(err, "error iterating over result rows")
	}

	return eventsList, nil
}
