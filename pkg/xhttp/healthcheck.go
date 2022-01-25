package xhttp

import (
	"net/http"
)

type HealthChecker func() error

type HealthCheck struct {
	uri      string
	checkers []HealthChecker
}

func NewHealthCheck(uri string, checkers ...HealthChecker) HealthCheck {
	return HealthCheck{uri, checkers}
}

func (h HealthCheck) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	for _, c := range h.checkers {
		err := c()
		if err != nil {
			WriteJsonError(w, err)
		}
	}

	WriteData(w, jsonOkString)
}

func (h HealthCheck) Handlers() Handlers {
	return Handlers{
		h.uri: h,
	}
}
