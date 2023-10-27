//go:build integration
// +build integration

package app

import (
	"context"
	"fmt"
	"github.com/gavv/httpexpect/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"protty/internal/infrastructure/app/mock"
	"protty/internal/infrastructure/config"
	"testing"
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
			logrusHook := mock.NewLogrusHook([]logrus.Level{logrus.InfoLevel}, 50)
			prottyApp.logger.Hooks.Add(logrusHook)
			go func() { assert.NoError(t, prottyApp.Start()) }()

			// waiting for the first info message from logger as a signal that the proxy is ready to handle requests
			<-logrusHook.EntryChan()

			// test
			e := httpexpect.Default(t, fmt.Sprintf("http://0.0.0.0:%d", cfg.LocalPort.Value))
			e.GET("/").Expect().Body().IsEqual(tt.want.responseBody)

			// stop protty and mock srv
			assert.NoError(t, prottyApp.Stop(context.Background()))
			assert.NoError(t, targetSrvMock.Shutdown(context.Background()))
		})
	}
}
