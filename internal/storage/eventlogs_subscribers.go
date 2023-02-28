package storage

import (
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"go.uber.org/zap"
)

func (storage *Storage) GetEventlogsSubscriber(subscriberID string) (*types.EventlogSubscriber, error) {
	query := `SELECT subscriber_id, log_id, offset, updated FROM eventlog_subscribers WHERE subscriber_id = :subscriber_id`
	params := struct {
		SubscriberID string `db:"subscriber_id"`
	}{
		SubscriberID: subscriberID,
	}

	rows, err := storage.db.NamedQuery(query, params)
	if err != nil {
		return nil, xerror.EStorageError("can't get eventlog subscriber by id", err, zap.String("subscriber_id", subscriberID))
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrNotFound
	}

	var eventlogSubscriber types.EventlogSubscriber
	err = rows.StructScan(&eventlogSubscriber)
	if err != nil {
		return nil, xerror.EStorageError("can't parse eventlog subscriber data", err, zap.String("subscriber_id", subscriberID))
	}

	return &eventlogSubscriber, nil
}

func (storage *Storage) PutEventlogsSubscriber(subscriber *types.EventlogSubscriber) error {
	now := xtime.Now()
	subscriber.Updated = &now
	query := `
		INSERT INTO eventlog_subscribers(subscriber_id, log_id, offset, updated) 
		VALUES(:subscriber_id, :log_id, :offset, :updated)
		ON CONFLICT(subscriber_id) 
		DO UPDATE SET log_id=excluded.log_id, offset=excluded.offset, updated=excluded.updated
`
	_, err := storage.db.NamedExec(query, subscriber)
	if err != nil {
		return xerror.EStorageError("can't put eventlog subscriber data", err, zap.Any("eventlog_subscriber", subscriber))
	}
	return nil
}

func (storage *Storage) DeleteEventlogsSubscriber(subscriberID string) error {
	query := `DELETE eventlog_subscribers WHERE subscriber_id = :subscriber_id`
	params := struct {
		SubscriberID string `db:"subscriber_id"`
	}{
		SubscriberID: subscriberID,
	}
	_, err := storage.db.NamedExec(query, params)
	if err != nil {
		return xerror.EStorageError("can't delete eventlog subscriber data", err, zap.String("subscriber_id", subscriberID))
	}
	return nil
}
