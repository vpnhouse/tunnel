
-- +migrate Up
-- +migrate StatementBegin
ALTER TABLE "peers" ADD column "activity" INTEGER;
ALTER TABLE "peers" ADD column "upstream" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "peers" ADD column "downstream" INTEGER NOT NULL DEFAULT 0;
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
ALTER TABLE "peers" DROP column "activity";
ALTER TABLE "peers" DROP column "upstream";
ALTER TABLE "peers" DROP column "downstream";
-- +migrate StatementEnd
