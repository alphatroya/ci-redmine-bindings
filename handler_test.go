package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseHeaders(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	rw := httptest.NewRecorder()
	handler(rw, req)
	if rw.Header().Get("Content-Type") != "application/json" {
		t.Error("response headers is not application/json")
	}
}

func TestRequestHeaders(t *testing.T) {
	cases := []struct {
		requestHeaders     map[string]string
		requestBody        io.Reader
		responseStatusCode int
		responseMessage    string
	}{
		{
			map[string]string{
				"Bitrise-Event-Type": "build/started",
			},
			nil,
			http.StatusOK,
			"Skipping done transition: build status is not finished",
		},
		{
			map[string]string{
				"Bitrise-Event-Type": "build/finished",
			},
			badBody{},
			http.StatusBadRequest,
			"Received wrong request data payload",
		},
		{
			map[string]string{
				"Bitrise-Event-Type": "build/finished",
			},
			newMockBody(``),
			http.StatusBadRequest,
			"Can't decode request payload json data",
		},
		{
			map[string]string{
				"Bitrise-Event-Type": "build/finished",
			},
			newMockBody(`{"build_triggered_workflow":"external", "build_status":1, "build_number":12}`),
			http.StatusOK,
			"Skipping done transition: build workflow is not internal",
		},
		{
			map[string]string{
				"Bitrise-Event-Type": "build/finished",
			},
			newMockBody(`{"build_triggered_workflow":"internal", "build_status":0, "build_number":12}`),
			http.StatusOK,
			"Skipping done transition: build status is not success",
		},
	}

	for i, mock := range cases {
		req, _ := http.NewRequest(http.MethodGet, "", mock.requestBody)
		for key, value := range mock.requestHeaders {
			req.Header.Set(key, value)
		}
		rw := httptest.NewRecorder()
		handler(rw, req)

		response := new(HookResponse)
		json.Unmarshal(rw.Body.Bytes(), &response)

		if response.Message != mock.responseMessage {
			t.Errorf("case %d: wrong message received, got %s, expected %s", i, response.Message, mock.responseMessage)
		}

		if rw.Code != mock.responseStatusCode {
			t.Errorf("case %d: wrong response code received, got %d, expected %d", i, rw.Code, mock.responseStatusCode)
		}
	}
}

type badBody struct {
}

func (b badBody) Read(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

type mockBody struct {
	p string
}

func newMockBody(p string) mockBody {
	return mockBody{p}
}

func (m mockBody) Read(p []byte) (n int, err error) {
	for i, b := range []byte(m.p) {
		p[i] = b
	}
	return len(m.p), io.EOF
}
