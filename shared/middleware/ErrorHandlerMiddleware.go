package shared_middleware

import (
	"errors"
	"net/http"
	"tikube-backend/shared/http_error"
	"tikube-backend/shared/utils"
	"time"
)

func ErrorHandlerMiddleware(handler utils.HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			path := r.URL.Path
			method := r.Method
			// Check if the error is of type HTTPError
			var httpErr *http_error.HTTPError
			if errors.As(err, &httpErr) {
				// Construct and send the JSON response with the status code and message from the error
				_ = utils.JSONResponse(w, httpErr.StatusCode, map[string]interface{}{"message": httpErr.Message, "statusCode": httpErr.StatusCode, "path": path, "method": method, "timestamp": time.Now().UTC()})
			} else {
				// For non-HTTPError, send a generic server error
				_ = utils.JSONResponse(w, http.StatusInternalServerError, map[string]interface{}{"message": "Internal Server Error", "statusCode": 500, "path": path, "method": method, "timestamp": time.Now().UTC()})
			}
		}
	}
}
