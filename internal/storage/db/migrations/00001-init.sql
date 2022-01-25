
-- +migrate Up
-- +migrate StatementBegin
CREATE TABLE IF NOT EXISTS peers (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    label           VARCHAR(128),
    type            INTEGER NOT NULL,
    wireguard_key   VARCHAR(44),
    ipv4            VARCHAR(15) NOT NULL,
    created         INTEGER NOT NULL,
    updated         INTEGER NOT NULL,
    expires         INTEGER,
    claims          TEXT,
    user_id         VARCHAR(256),
    installation_id VARCHAR(36),
    session_id      VARCHAR(36)
);

CREATE INDEX IF NOT EXISTS peers_id ON peers(id);
CREATE INDEX IF NOT EXISTS peers_label ON peers(label);
CREATE INDEX IF NOT EXISTS peers_type ON peers(type);
CREATE UNIQUE INDEX IF NOT EXISTS peers_wireguard_key ON peers(wireguard_key);
CREATE UNIQUE INDEX IF NOT EXISTS peers_ipv4 ON peers(ipv4);
CREATE UNIQUE INDEX IF NOT EXISTS peers_identifiers ON peers(user_id, installation_id);
CREATE INDEX IF NOT EXISTS peers_session_id ON peers(session_id);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE peers;
-- +migrate StatementEnd
