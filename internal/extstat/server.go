/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package extstat

import (
	"bytes"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiPathInstall   = "/api/install"
	apiPathHeartbeat = "/api/heartbeat"
)

//go:embed migrations
var migrations embed.FS

//go:embed templates/index.html
var indexTmpl []byte

type server struct {
	db                 *sql.DB
	username, password string
}

func NewServer(username, password string) *server {
	sqlDB, err := sql.Open("sqlite3", "extstat-data/db.sqlite3")
	if err != nil {
		panic(err)
	}

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "migrations",
	}

	if _, err := migrate.Exec(sqlDB, "sqlite3", migrations, migrate.Up); err != nil {
		panic(err)
	}

	return &server{
		db:       sqlDB,
		username: username,
		password: password,
	}
}

func (s *server) Run(la string) {
	// TODO(nikonov): in-app SSL support?
	hs := xhttp.NewDefault()
	hs.Router().Get("/", s.handleIndexRequest)
	hs.Router().Post(apiPathInstall, s.handleInstallRequest)
	hs.Router().Post(apiPathHeartbeat, s.handleHeartbeatRequest)
	if err := hs.Run(la); err != nil {
		panic(err)
	}
}

func (s *server) handleInstallRequest(w http.ResponseWriter, r *http.Request) {
	defer s.postRequest(r)

	body, err := s.readRequestBody(r)
	if err != nil {
		zap.L().Warn("failed to read request body", zap.Error(err), zap.String("instance_id", body.InstanceID))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.storeInstallRequest(body); err != nil {
		zap.L().Warn("failed to store the install request", zap.Error(err), zap.String("instance_id", body.InstanceID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	zap.L().Info("got new install request", zap.Any("req", body))
}

func (s *server) handleHeartbeatRequest(w http.ResponseWriter, r *http.Request) {
	defer s.postRequest(r)

	body, err := s.readRequestBody(r)
	if err != nil {
		zap.L().Warn("failed to read request body", zap.Error(err), zap.String("instance_id", body.InstanceID))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.storeHeartbeatRequest(body); err != nil {
		zap.L().Warn("failed to store the heartbeat request", zap.Error(err), zap.String("instance_id", body.InstanceID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	zap.L().Debug("got new heartbeat request", zap.Any("req", body))
}

func (s *server) storeInstallRequest(req installRequest) error {
	_, err := s.db.Exec(`insert into installs(installid, created_at, version, gitcommit) values (?, ?, ?, ?)
			on conflict(installid) do update set repeat = repeat+1, version=excluded.version, gitcommit=excluded.gitcommit;`,
		req.InstanceID, time.Now().UTC(), req.VersionTag, req.VersionCommit)
	return err
}

func (s *server) storeHeartbeatRequest(req installRequest) error {
	_, err := s.db.Exec(`insert into heartz(installid, created_at, version, gitcommit) values (?, ?, ?, ?)`,
		req.InstanceID, time.Now().UTC(), req.VersionTag, req.VersionCommit)
	return err
}

func (s *server) readRequestBody(r *http.Request) (installRequest, error) {
	body := installRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return installRequest{}, fmt.Errorf("failed to decode request body: %v", err)
	}
	if len(body.InstanceID) == 0 {
		return installRequest{}, fmt.Errorf("got empty instance_id")
	}
	if len(body.VersionCommit) == 0 {
		zap.L().Warn("no commit info in the request")
		body.VersionCommit = "head"
	}
	if len(body.VersionTag) == 0 {
		zap.L().Warn("no version info in the request")
		body.VersionTag = "latest"
	}
	return body, nil
}

func (s *server) postRequest(r *http.Request) {
	// todo: apply rate limiter here via the xdp or nftables
	// r.RemoteAddr
}

func (s *server) handleIndexRequest(w http.ResponseWriter, r *http.Request) {
	if ok := s.handleAuth(r); !ok {
		// force browser to show the standard auth prompt
		w.Header().Set("www-authenticate", "basic")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	q := `select * from
  (select count(*) as count from installs where created_at > datetime('now', '-1 day')) daily,
  (select count(*) as count from installs where created_at > datetime('now', '-7 days')) weekly,
  (select count(*) as count from installs where created_at > datetime('now', '-31 days')) monthly,
  (select count(*) as count from installs) total;`

	row := s.db.QueryRow(q)
	var daily, weekly, montly, total int64
	if err := row.Scan(&daily, &weekly, &montly, &total); err != nil {
		zap.L().Warn("failed to scan counters", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	q = `select count(*) from (select 1 from heartz where created_at > datetime('now', '-1 days') group by installid);`
	row = s.db.QueryRow(q)
	var liveInstalls int64
	if err := row.Scan(&liveInstalls); err != nil {
		zap.L().Warn("failed to scan live installs", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl := s.getTemplate()
	tmpl = s.render(tmpl, daily, weekly, montly, total, liveInstalls)
	w.Header().Set("content-type", "text/html")
	w.Write(tmpl)
}

func (s *server) getTemplate() []byte {
	// for debug purposes
	// bs, err := os.ReadFile("internal/extstat/templates/index.html")
	return indexTmpl
}

func i64tobs(i int64) []byte {
	return []byte(strconv.Itoa(int(i)))
}

func (s *server) render(temp []byte, d, w, m, t, li int64) []byte {
	temp = bytes.Replace(temp, []byte("$INSTALL_24H$"), i64tobs(d), -1)
	temp = bytes.Replace(temp, []byte("$INSTALL_WEEK$"), i64tobs(w), -1)
	temp = bytes.Replace(temp, []byte("$INSTALL_MONTH$"), i64tobs(m), -1)
	temp = bytes.Replace(temp, []byte("$INSTALL_TOTAL$"), i64tobs(t), -1)
	temp = bytes.Replace(temp, []byte("$INSTALL_ALIVE$"), i64tobs(li), -1)
	return temp
}

func (s *server) handleAuth(r *http.Request) bool {
	u, p, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return u == s.username && p == s.password
}
