package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogRequestPayload(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		rawQuery      string
		host          string
		wantMethod    string
		wantPath      string
		wantQuery     string
		wantMsgSuffix string
	}{
		{
			name:          "GET request without query params",
			method:        "GET",
			path:          "/api/resource",
			rawQuery:      "",
			host:          "example.com",
			wantMethod:    "GET",
			wantPath:      "/api/resource",
			wantQuery:     "",
			wantMsgSuffix: "example.com",
		},
		{
			name:          "GET request with single query param",
			method:        "GET",
			path:          "/api/resource",
			rawQuery:      "id=123",
			host:          "example.com",
			wantMethod:    "GET",
			wantPath:      "/api/resource",
			wantQuery:     "id=123",
			wantMsgSuffix: "example.com",
		},
		{
			name:          "GET request with multiple query params",
			method:        "GET",
			path:          "/itxrest/2/catalog/store/35009503/category/1835010/product/58227962/detail",
			rawQuery:      "lang=en&currency=USD&fields=all",
			host:          "api.example.com",
			wantMethod:    "GET",
			wantPath:      "/itxrest/2/catalog/store/35009503/category/1835010/product/58227962/detail",
			wantQuery:     "lang=en&currency=USD&fields=all",
			wantMsgSuffix: "api.example.com",
		},
		{
			name:          "POST request with query params",
			method:        "POST",
			path:          "/api/data",
			rawQuery:      "action=create",
			host:          "example.com",
			wantMethod:    "POST",
			wantPath:      "/api/data",
			wantQuery:     "action=create",
			wantMsgSuffix: "example.com",
		},
		{
			name:          "request with encoded query params",
			method:        "GET",
			path:          "/search",
			rawQuery:      "q=hello+world&filter=name%3Dfoo",
			host:          "example.com",
			wantMethod:    "GET",
			wantPath:      "/search",
			wantQuery:     "q=hello+world&filter=name%3Dfoo",
			wantMsgSuffix: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.InfoLevel)

			svc := NewReverseProxyService(logger)

			req := &http.Request{
				Method: tt.method,
				URL: &url.URL{
					Path:     tt.path,
					RawQuery: tt.rawQuery,
					Host:     tt.host,
				},
			}

			svc.logRequestPayload(req)

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			require.NoError(t, err, "log output should be valid JSON")

			assert.Equal(t, tt.wantMethod, logEntry["method"])
			assert.Equal(t, tt.wantPath, logEntry["path"])
			assert.Equal(t, tt.wantQuery, logEntry["query"])
			assert.Contains(t, logEntry["msg"], tt.wantMsgSuffix)
			assert.Equal(t, "info", logEntry["level"])
		})
	}
}
