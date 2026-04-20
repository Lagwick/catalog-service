package httph

import (
	"encoding/json"
	"net/http"
)

func EncodeJSON(w http.ResponseWriter, data interface{}) error {
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func DecodeJSON(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return err
	}
	return nil
}
