-- +migrate Up
-- +migrate StatementBegin
UPDATE metrics SET name='upstream_wireguard' WHERE name='upstream';
UPDATE metrics SET name='downstream_wireguard' WHERE name='downstream';
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
UPDATE metrics SET name='upstream' WHERE name='upstream_wireguard';
UPDATE metrics SET name='downstream' WHERE name='downstream_wireguard';
-- +migrate StatementEnd