package httputils

import (
	"encoding/json"
	"io"
	"net/http"
)

func SendTextResponse(w http.ResponseWriter, code int, text string) error {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(code)

	if _, err := io.WriteString(w, text); err != nil {
		return err
	}

	return nil
}

func SendJSONResponse(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set("content-type", "application/json")

	jsonData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.WriteHeader(code)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}

func SendJSONErrorResponse(w http.ResponseWriter, code int, error string) error {
	return SendJSONResponse(w, code, map[string]string{
		"error": error,
	})
}

func SendRedirectResponse(w http.ResponseWriter, location string) {
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func SendStatusCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}
