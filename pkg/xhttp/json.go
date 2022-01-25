package xhttp

import (
	"encoding/json"
)

func JSONMarshal(v interface{}) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}

	return b, nil
}
