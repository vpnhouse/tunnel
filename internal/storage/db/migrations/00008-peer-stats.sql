
-- +migrate Up
-- +migrate StatementBegin
ALTER TABLE "peers" ADD column "upstream" INTEGER DEFAULT 0;
ALTER TABLE "peers" ADD column "downstream" INTEGER DEFAULT 0;
ALTER TABLE "peers" ADD column "activity" INTEGER;
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
ALTER TABLE "peers" DROP column "upstream";
ALTER TABLE "peers" DROP column "downstream";
ALTER TABLE "peers" DROP column "activity";
-- +migrate StatementEnd
