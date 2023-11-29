//go:build integration
// +build integration

package app

import (
	"context"
	"fmt"
	"github.com/gavv/httpexpect/v2"
	"github.com/mgerasimchuk/protty/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestStartCommand(t *testing.T) {
	targetPort := ":8081"
	targetURI := "http://127.0.0.1" + targetPort

	prottyStart := []string{"protty", "start", "--remote-uri", targetURI}

	type args struct {
		targetPath         string
		targetResponseBody string

		prottyFlags []string
		prottyEnvs  []string // TODO

		requestHeaders []string // TODO
	}
	type want struct {
		responseBody string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"Flags configuration Remote uri",
			args{targetPath: "/", targetResponseBody: "ok", prottyFlags: prottyStart},
			want{responseBody: "ok"},
		},
		{
			"Flags configuration SED expression",
			args{targetPath: "/", targetResponseBody: "ok", prottyFlags: append(prottyStart, "--transform-response-body-sed", "s|ok|changed|g")},
			want{responseBody: "changed"},
		},
		{
			"Flags configuration JQ expression",
			args{targetPath: "/", targetResponseBody: `{"code": 100, "message": "message body"}`, prottyFlags: append(prottyStart, "--transform-response-body-jq", ".message")},
			want{responseBody: "message body"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// target server mock
			mux := http.NewServeMux()
			mux.HandleFunc(tt.args.targetPath, func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte(tt.args.targetResponseBody))
			})
			targetSrvMock := &http.Server{Addr: targetPort, Handler: mux}
			go func() {
				if err := targetSrvMock.ListenAndServe(); err != nil {
					assert.ErrorIs(t, err, http.ErrServerClosed)
				}
			}()

			// protty
			os.Args = tt.args.prottyFlags
			cfg := config.GetStartCommandConfig()
			prottyApp := NewProttyApp(cfg)
			prottyApp.logger.SetOutput(io.Discard)
			go func() { assert.NoError(t, prottyApp.Start()) }()

			// waiting for the first info message from logger as a signal that the proxy is ready to handle requests
			for i := 0; i < 5; i++ {
				_, err := http.Get("http://0.0.0.0:80")
				if err == nil {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}

			// test
			e := httpexpect.Default(t, fmt.Sprintf("http://0.0.0.0:%d", cfg.LocalPort.Value))
			e.GET("/").Expect().Body().IsEqual(tt.want.responseBody)

			// stop protty and mock srv
			assert.NoError(t, prottyApp.Stop(context.Background()))
			assert.NoError(t, targetSrvMock.Shutdown(context.Background()))
		})
	}
}
