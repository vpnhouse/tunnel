package xhttp

import (
	"net/http"
)

type RequestInfoStruct struct {
	URI  string
	Host string
}

func RequestInfo(r *http.Request) *RequestInfoStruct {
	return &RequestInfoStruct{
		URI:  r.RequestURI,
		Host: r.Host,
	}
}
