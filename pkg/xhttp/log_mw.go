// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xhttp

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(s int) {
	r.status = s
	r.ResponseWriter.WriteHeader(s)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		wr := &recorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wr, r)

		zap.L().Info("request info",
			zap.String("host", r.Host),
			zap.String("method", r.Method),
			zap.String("path", r.RequestURI),
			zap.Int("status", wr.status),
			zap.Duration("t", time.Since(start)),
		)
	})
}

func (w recorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}
