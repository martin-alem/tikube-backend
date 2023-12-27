package shared_middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"tikube-backend/shared/http_error"
	"tikube-backend/shared/utils"
)

func PayloadValidationMiddleware[T utils.ValidatableSchema](factory func() T) utils.Middleware {
	return func(next utils.HTTPHandler) utils.HTTPHandler {
		return func(w http.ResponseWriter, r *http.Request) error {

			r.Body = http.MaxBytesReader(w, r.Body, utils.MaxPayloadSize)
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields() // return an error if extra fields are present

			payloadSchema := factory()

			if err := decoder.Decode(&payloadSchema); err != nil {
				return http_error.BadRequest("Invalid payload: " + err.Error())
			}

			// Validate fields
			if err := payloadSchema.Validate(); err != nil {
				return http_error.BadRequest("Validation error: " + err.Error())
			}

			ctx := context.WithValue(r.Context(), utils.PayloadKey{}, payloadSchema)
			return next(w, r.WithContext(ctx))
		}
	}
}
