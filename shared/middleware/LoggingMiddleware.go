package shared_middleware

import (
	"fmt"
	"net/http"
	"tikube-backend/shared/utils"
	"time"
)

func LoggingMiddleware(next utils.HTTPHandler) utils.HTTPHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		start := time.Now()

		fmt.Printf("%s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)

		err := next(w, r)

		duration := time.Since(start)
		fmt.Printf("Completed in %v\n", duration)

		return err
	}
}
