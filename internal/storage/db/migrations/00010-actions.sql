-- +migrate Up
-- +migrate StatementBegin
CREATE TABLE IF NOT EXISTS actions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         VARCHAR(256),
    expires         INTEGER,
    rules_json      TEXT,
);

CREATE INDEX IF NOT EXISTS actions_id ON actions(id);
CREATE INDEX IF NOT EXISTS actions_user_id ON actions(user_id);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE actions;
-- +migrate StatementEnd