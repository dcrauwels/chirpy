package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {

	auth := headers.Get("Authorization")
	if auth == "" {
		return auth, fmt.Errorf("no authorization header value found")
	}
	return strings.TrimPrefix(auth, "ApiKey "), nil

}
