/*
 * Copyright 2021 The VPNHouse Authors. All rights reserved.
 * Use of this source code is governed by a AGPL-style
 * license that can be found in the LICENSE file.
 */

-- +migrate Up
-- +migrate StatementBegin
create table if not exists installs (
     id         integer primary key autoincrement,
     installid  varchar(50) not null,
     version    varchar(50),
     gitcommit  varchar(50),
     created_at integer not null,
     repeat     integer not null default 0
);
create index if not exists idx_installs_version ON installs(version);
create unique index if not exists uk_installs_iid on installs(installid);


create table if not exists heartz(
    id         integer primary key autoincrement,
    installid  varchar(50) not null ,
    version    varchar(50),
    gitcommit  varchar(50),
    created_at integer not null
);
create index if not exists idx_heartz_installid ON heartz(installid);
create index if not exists idx_heartz_version ON heartz(version);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE installs;
DROP TABLE heartz;
-- +migrate StatementEnd
