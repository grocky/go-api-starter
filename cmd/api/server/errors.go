package server

import (
	"fmt"
	"go-api-starter/cmd/api/response"
	"go-api-starter/internal/log"
	"go-api-starter/internal/validator"
	"net/http"
	"runtime/debug"
)

func ErrorMessage(w http.ResponseWriter, r *http.Request, status int, clientMessage string) {
	err := response.JSON(w, status, map[string]string{"error": clientMessage})
	if err != nil {
		logger := log.FromContext(r.Context())
		logger.Error("unable to marshal json response", "error", err, "clientMessage", clientMessage)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func ErrorMessageLog(w http.ResponseWriter, r *http.Request, status int, clientMessage string, e error) {
	logger := log.FromContext(r.Context())
	logger.Error(e.Error(), "debug", debug.Stack())
	ErrorMessage(w, r, status, clientMessage)
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	message := "The server encountered a problem and could not process your request"
	ErrorMessageLog(w, r, http.StatusInternalServerError, message, err)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	ErrorMessage(w, r, http.StatusNotFound, message)
	http.NotFoundHandler()
}

func NotFoundHandler() http.Handler { return http.HandlerFunc(NotFound) }

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	ErrorMessage(w, r, http.StatusMethodNotAllowed, message)
}

func MethodNotAllowedHandler() http.Handler { return http.HandlerFunc(MethodNotAllowed) }

func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	ErrorMessage(w, r, http.StatusBadRequest, err.Error())
}

func FailedValidation(w http.ResponseWriter, r *http.Request, v validator.Validator) {
	err := response.JSON(w, http.StatusUnprocessableEntity, v)
	if err != nil {
		Error(w, r, err)
	}
}

func InvalidAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", "Bearer")

	ErrorMessage(w, r, http.StatusUnauthorized, "Invalid authentication token")
}

func AuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	ErrorMessage(w, r, http.StatusUnauthorized, "You must be authenticated to access this resource")
}

func BasicAuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	message := "You must be authenticated to access this resource"
	ErrorMessage(w, r, http.StatusUnauthorized, message)
}
