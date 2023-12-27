package utils

import (
	"encoding/json"
	"net"
	"net/http"
)

func GetClientIP(req *http.Request) string {
	// Standard headers used by Amazon ELB, Heroku, and others.
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback to using the remote address from the request.
	// This will give the network IP, which might be a proxy.
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}

func SerializeKafkaMessage[T any](message T) (string, error) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func CreateSerializedLog(level LogLevel, source string, message string) string {
	log := CreateLogSchema{
		level,
		source,
		message,
	}
	d, _ := SerializeKafkaMessage[CreateLogSchema](log)
	return d
}
