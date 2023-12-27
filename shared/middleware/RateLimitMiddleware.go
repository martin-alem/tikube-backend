package shared_middleware

import (
	"github.com/go-redis/redis_rate/v10"
	"net/http"
	"strconv"
	"tikube-backend/shared/http_error"
	"tikube-backend/shared/utils"
	"time"
)

func RateLimitMiddleware(rateLimiter *redis_rate.Limiter, limit redis_rate.Limit) utils.Middleware {
	return func(next utils.HTTPHandler) utils.HTTPHandler {
		return func(w http.ResponseWriter, r *http.Request) error {

			clientIP := utils.GetClientIP(r)
			key := "rate_limit:" + clientIP

			res, err := rateLimiter.Allow(r.Context(), key, limit)
			if err != nil {
				return err
			}

			h := w.Header()
			h.Set("RateLimit-Remaining", strconv.Itoa(res.Remaining))

			if res.Allowed == 0 {
				seconds := int(res.RetryAfter / time.Second)
				h.Set("RateLimit-RetryAfter", strconv.Itoa(seconds))
				return http_error.TooManyRequests()
			}
			return next(w, r)
		}
	}
}
