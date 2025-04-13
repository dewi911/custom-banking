package models

import (
	"errors"
	"time"
)

type eventType string
type metadata map[string]any

func (et eventType) String() string {
	return string(et)
}

func NewEventTypeFromString(et string) (eventType, error) {
	switch et {
	case AccountCreatedEvent.String():
		return AccountCreatedEvent, nil
	case AccountDeletedEvent.String():
		return AccountDeletedEvent, nil
	case AccountBlockedEvent.String():
		return AccountBlockedEvent, nil
	case AccountUnblockedEvent.String():
		return AccountUnblockedEvent, nil
	case CardCreatedEvent.String():
		return CardCreatedEvent, nil
	case UserBlockedEvent.String():
		return UserBlockedEvent, nil
	case UserUnblockedEvent.String():
		return UserUnblockedEvent, nil
	case WithdrawalEvent.String():
		return WithdrawalEvent, nil
	case DepositEvent.String():
		return DepositEvent, nil
	default:
		return "", errors.New("unsupported event type")
	}
}

const (
	AccountCreatedEvent   eventType = "ACCOUNT_CREATED"
	AccountDeletedEvent   eventType = "ACCOUNT_DELETED"
	AccountBlockedEvent   eventType = "ACCOUNT_BLOCKED"
	AccountUnblockedEvent eventType = "ACCOUNT_UNBLOCKED"
	CardCreatedEvent      eventType = "CARD_CREATED"
	UserBlockedEvent      eventType = "USER_BLOCKED"
	UserUnblockedEvent    eventType = "USER_UNBLOCKED"
	WithdrawalEvent       eventType = "WITHDRAWAL"
	DepositEvent          eventType = "DEPOSIT"
)

type Event struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`
	Type     eventType `json:"type"`
	Message  string    `json:"message"`
	Metadata metadata  `json:"metadata"`
	DateTime time.Time `json:"date_time"`
}
