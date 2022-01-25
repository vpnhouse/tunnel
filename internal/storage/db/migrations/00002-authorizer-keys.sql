-- +migrate Up
-- +migrate StatementBegin
CREATE TABLE IF NOT EXISTS authorizer_keys (
    id     char(36) PRIMARY KEY,
    source char(100) not null,
    key    text not null
);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE IF EXISTS authorizer_keys;
-- +migrate StatementEnd
