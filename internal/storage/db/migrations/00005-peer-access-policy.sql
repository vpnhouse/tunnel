-- +migrate Up
-- +migrate StatementBegin
alter table "peers" add column "net_access_policy" int default 0;
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
alter table "peers" drop column "net_access_policy";
-- +migrate StatementEnd
