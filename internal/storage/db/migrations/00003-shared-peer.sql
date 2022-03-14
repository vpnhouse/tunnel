-- +migrate Up
-- +migrate StatementBegin
alter table "peers" add column "sharing_key" varchar(125) default '';
alter table "peers" add column "sharing_key_expiration" integer default 0;

-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
alter table "peers" drop column "sharing_key";
alter table "peers" drop column "sharing_key_expiration";
-- +migrate StatementEnd
