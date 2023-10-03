package http_utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendStatusCode(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{
			name: "Send success status code",
			code: http.StatusOK,
		},
		{
			name: "Send error status code",
			code: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SendStatusCode(w, test.code)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.code, res.StatusCode)
		})
	}
}

func TestSendTextResponse(t *testing.T) {
	t.Run("Send text response", func(t *testing.T) {
		text := "Hello"
		w := httptest.NewRecorder()
		err := SendTextResponse(w, http.StatusOK, text)
		require.NoError(t, err)

		res := w.Result()
		require.Equal(t, http.StatusOK, res.StatusCode)
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, "text/plain", res.Header.Get("content-type"))
		assert.Equal(t, text, string(body))
	})
}

func TestSendJSONResponse(t *testing.T) {
	t.Run("Send json response", func(t *testing.T) {
		data := map[string]string{
			"id": "test-id",
		}

		jsonData, err := json.Marshal(data)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		err = SendJSONResponse(w, http.StatusCreated, data)
		require.NoError(t, err)

		res := w.Result()
		require.Equal(t, http.StatusCreated, res.StatusCode)
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, "application/json", res.Header.Get("content-type"))
		assert.JSONEq(t, string(jsonData), string(body))
	})
}

func TestSendRedirectResponse(t *testing.T) {
	t.Run("Send redirect response", func(t *testing.T) {
		url := "https://practicum.yandex.ru"
		w := httptest.NewRecorder()
		SendRedirectResponse(w, url)

		res := w.Result()
		require.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
		defer res.Body.Close()

		assert.Equal(t, url, res.Header.Get("Location"))
	})
}
