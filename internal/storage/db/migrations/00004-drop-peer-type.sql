-- +migrate Up
-- +migrate StatementBegin
drop index if exists peers_type;
alter table "peers" drop column "type";
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
alter table "peers" add column "type" integer not null default 0;
CREATE INDEX peers_type ON peers(type);
-- +migrate StatementEnd
