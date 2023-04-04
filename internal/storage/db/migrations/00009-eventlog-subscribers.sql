-- +migrate Up
-- +migrate StatementBegin
CREATE TABLE IF NOT EXISTS eventlog_subscribers (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    subscriber_id   VARCHAR(64) NOT NULL DEFAULT "",
    log_id          VARCHAR(64) NOT NULL DEFAULT "",
    offset          INTEGER NOT NULL DEFAULT 0,
    updated         INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS eventlog_subscribers_subscriber_id ON eventlog_subscribers(subscriber_id);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE eventlog_subscribers;
-- +migrate StatementEnd