package util

import (
	"net/http"
	"strings"
)

func IsFilteredHttpRoute(r *http.Request) bool {
	if strings.Contains(r.RequestURI, "health") ||
		strings.Contains(r.RequestURI, "debug") ||
		strings.Contains(r.RequestURI, "metric") {
		return true
	}
	return false
}
