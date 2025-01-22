package httptoreq

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestParseHTTPFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected *http.Request
		hasError bool
	}{
		{
			name:    "Single GET request",
			content: `GET https://example.com`,
			expected: func() *http.Request {
				req, _ := http.NewRequest("GET", "https://example.com", nil)
				return req
			}(),
			hasError: false,
		},
		{
			name: "Multiple requests with body",
			content: `POST https://example.com/resource
				Content-Type: application/json

				{"key": "value"}
				###
				DELETE https://example.com/resource/123`,

			expected: func() *http.Request {
				body := bytes.NewBufferString(`{"key": "value"}`)
				req, _ := http.NewRequest("POST", "https://example.com/resource", body)
				req.Header.Add("Content-Type", "application/json")
				return req
			}(),
			hasError: false,
		},
		{
			name: "Malformed request (missing method)",
			content: `https://example.com
				Authorization: Bearer <token>`,
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetOne([]byte(tt.content))
			if tt.hasError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check result for a single request
			if !compareRequests(result, tt.expected) {
				t.Errorf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func compareRequests(req1, req2 *http.Request) bool {
	if req1.Method != req2.Method || req1.URL.String() != req2.URL.String() {
		return false
	}
	if !reflect.DeepEqual(req1.Header, req2.Header) {
		return false
	}

	var body1, body2 []byte
	var err error

	if req1.Body != nil {
		body1, err = io.ReadAll(req1.Body)
		if err != nil {
			return false
		}
		req1.Body = io.NopCloser(bytes.NewBuffer(body1))
	}

	if req2.Body != nil {
		body2, err = io.ReadAll(req2.Body)
		if err != nil {
			return false
		}
		req2.Body = io.NopCloser(bytes.NewBuffer(body2))
	}
	return bytes.Equal(body1, body2)
}
