package utils

import (
	"encoding/json"
	"net/http"
	"tikube-backend/shared/http_error"
)

// JSONResponse sends a JSON response with a given status code.
func JSONResponse(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			return http_error.InternalServerError()
		}
	}
	return nil
}

// JSONDecode decodes a JSON request into a struct.
func JSONDecode(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(dst)
}
