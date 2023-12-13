package cors

import (
	"net/http"
	"os"
	"strconv"
	"strings"
)

type CorsOptions struct {
	AllowedMethods, AllowedHeaders, AllowedOrigins []string
	AllowedCredentials                             bool
}

func retriveOriginEnv(name string) []string {
	return strings.Split(os.Getenv(name), " ")
}

func NewCors() CorsOptions {
	allowedOrigins := []string{"*"}
	if envOrigin := retriveOriginEnv("ALLOWED_ORIGINS"); envOrigin[0] != "" {
		allowedOrigins = envOrigin
	}
	allowedMethods := []string{http.MethodHead,
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}
	if envMethod := retriveOriginEnv("ALLOWED_METHODS"); envMethod[0] == "" {
		allowedMethods = []string{"POST", "GET", "OPTIONS"}
	}
	allowedHeaders := []string{"*"}
	if envHeaders := retriveOriginEnv("ALLOWED_HEADERS"); envHeaders[0] == "" {
		allowedHeaders = []string{"Accept", "Authorization", "Origin", "Content-Type"}
	}
	allowedCredentials := true
	var err error
	if envCredentials := os.Getenv("ALLOWED_CREDENTIALS"); envCredentials != "" {
		allowedCredentials, err = strconv.ParseBool(envCredentials)
		if err != nil {
			panic("cannot parse  ALLOWED_CREDENTIALS env to boolean")
		}
	}
	c := CorsOptions{
		AllowedMethods:     allowedMethods,
		AllowedHeaders:     allowedHeaders,
		AllowedOrigins:     allowedOrigins,
		AllowedCredentials: allowedCredentials,
	}
	return c
}
