package http_error

// HTTPError struct represents an error with an associated HTTP status code.
type HTTPError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface.
func (e HTTPError) Error() string {
	return e.Message
}

// BadRequest returns a 400 Bad Request error.
func BadRequest(messages ...string) *HTTPError {
	message := "Bad Request"

	if len(messages) > 0 {
		message = messages[0]
	}

	return &HTTPError{400, message}
}

// Unauthorized returns a 401 Unauthorized error.
func Unauthorized(messages ...string) *HTTPError {
	message := "Unauthorized"

	if len(messages) > 0 {
		message = messages[0]
	}

	return &HTTPError{401, message}
}

// Forbidden returns a 403 Forbidden error.
func Forbidden(messages ...string) *HTTPError {
	message := "Forbidden"

	if len(messages) > 0 {
		message = messages[0]
	}
	return &HTTPError{403, message}
}

// NotFound returns a 404 Not Found error.
func NotFound(messages ...string) *HTTPError {
	message := "Not Found"

	if len(messages) > 0 {
		message = messages[0]
	}
	return &HTTPError{404, message}
}

// PayloadTooLarge returns a 413 Payload Too Large error.
func PayloadTooLarge(messages ...string) *HTTPError {
	message := "Payload Too Large"

	if len(messages) > 0 {
		message = messages[0]
	}
	return &HTTPError{413, message}
}

// TooManyRequests returns a 429 Too Many Requests error.
func TooManyRequests(messages ...string) *HTTPError {
	message := "Too Many Request"

	if len(messages) > 0 {
		message = messages[0]
	}
	return &HTTPError{429, message}
}

// RequestHeaderFieldsTooLarge returns a 431 Request Header Fields Too Large error.
func RequestHeaderFieldsTooLarge(messages ...string) *HTTPError {
	message := "Request Header  Fields Too Large"

	if len(messages) > 0 {
		message = messages[0]
	}
	return &HTTPError{431, message}
}

// InternalServerError returns a 500 Internal Server Error.
func InternalServerError(messages ...string) *HTTPError {
	message := "Internal Server Error"

	if len(messages) > 0 {
		message = messages[0]
	}
	return &HTTPError{500, message}
}
