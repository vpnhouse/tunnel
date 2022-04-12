-- +migrate Up
-- +migrate StatementBegin
alter table "peers" add column "net_rate_limit" int default 0;
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
alter table "peers" drop column "net_rate_limit";
-- +migrate StatementEnd
