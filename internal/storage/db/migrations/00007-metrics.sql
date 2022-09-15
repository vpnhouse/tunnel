
-- +migrate Up
-- +migrate StatementBegin
CREATE TABLE IF NOT EXISTS metrics (
    name            VARCHAR(32),
    value           INTEGER NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS metrics_name ON metrics(name);
INSERT OR IGNORE INTO metrics (name, value) VALUES ("upstream", 0);
INSERT OR IGNORE INTO metrics (name, value) VALUES ("downstream", 0);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE metrics;
-- +migrate StatementEnd
