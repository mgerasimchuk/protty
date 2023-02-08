package service

import (
	"fmt"
	"github.com/graze/go-throttled"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/httputil"
	"net/url"
	"protty/internal/infrastructure/config"
	"strconv"
	"strings"
)

type ReverseProxyService struct {
	reverseProxies map[string]*httputil.ReverseProxy
	cfg            *config.StartCommandConfig
	logger         *logrus.Logger
}

func NewReverseProxyService(logger *logrus.Logger) *ReverseProxyService {
	s := &ReverseProxyService{logger: logger, reverseProxies: map[string]*httputil.ReverseProxy{}}
	return s
}

func (s *ReverseProxyService) Start(cfg *config.StartCommandConfig) error {
	s.cfg = cfg

	http.HandleFunc("/", s.handleRequestAndRedirect)

	s.logger.Infof("Start listen proxy on :%d port with config: %+v", s.cfg.LocalPort.Value, s.cfg)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.cfg.LocalPort.Value), nil)
}

// Serve a reverse proxy for a given url
func (s *ReverseProxyService) serveReverseProxy(res http.ResponseWriter, req *http.Request) {
	cfg := s.getOverrideConfig(req)

	host := strings.ReplaceAll(strings.ReplaceAll(cfg.RemoteURI.Value, "https://", ""), "http://", "")
	req.URL.Host = host
	req.Host = host

	reverseProxy := s.getReverseProxyByParams(cfg)
	reverseProxy.ServeHTTP(res, req)
}

func (s *ReverseProxyService) logRequestPayload(req *http.Request) {
	s.logger.WithField("method", req.Method).WithField("path", req.URL.Path).Infof("Send request to %s", req.URL.Host)
	// TODO add tracing log with body and other params like headers
}

func (s *ReverseProxyService) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	s.serveReverseProxy(res, req)
	s.logRequestPayload(req)
}

func (s *ReverseProxyService) getOverrideConfig(req *http.Request) *config.StartCommandConfig {
	cfg := *s.cfg

	// TODO Refactor to dynamic configuring based on the cfg.Range(f func()) function
	if v := req.Header.Get(cfg.RemoteURI.GetHeaderName()); v != "" {
		cfg.RemoteURI.Value = v
		s.logger.Debugf("RemoteURI has been changed based on request's header")
	}
	if v := req.Header.Get(cfg.ThrottleRateLimit.GetHeaderName()); v != "" {
		throttleRateLimit, err := strconv.ParseFloat(v, 64)
		if err == nil {
			cfg.ThrottleRateLimit.Value = throttleRateLimit
			s.logger.Debugf("ThrottleRateLimit has been changed based on request's header")
		} else {
			s.logger.Errorf("Parsing %s error: %s", cfg.ThrottleRateLimit.GetHeaderName(), err)
		}
	}
	if v := req.Header.Get(cfg.ThrottleHost.GetHeaderName()); v != "" {
		cfg.ThrottleHost.Value = v
		s.logger.Debugf("ThrottleHost has been changed based on request's header")
	}

	if err := cfg.Validate(); err != nil {
		s.logger.Errorf("Validation override config error: %s. Reverting to original config", err)
		cfg = *s.cfg
	}

	return &cfg
}

func (s *ReverseProxyService) getReverseProxyByParams(cfg *config.StartCommandConfig) *httputil.ReverseProxy {
	var reverseProxy *httputil.ReverseProxy
	var ok bool
	reverseProxyKey := fmt.Sprintf("%s-%f-%s", cfg.RemoteURI, cfg.ThrottleRateLimit, cfg.ThrottleHost)

	if reverseProxy, ok = s.reverseProxies[reverseProxyKey]; !ok {
		remoteURL, _ := url.Parse(cfg.RemoteURI.Value)
		reverseProxy = httputil.NewSingleHostReverseProxy(remoteURL)
		if cfg.ThrottleRateLimit.Value != 0 {
			// Another way to throttle requests on the handler side: https://github.com/go-chi/chi/blob/878319e482623b6e9c5787147e5b481f8879c49e/_examples/limits/main.go#L75
			reverseProxy.Transport = throttled.NewTransport(http.DefaultTransport, rate.NewLimiter(rate.Limit(cfg.ThrottleRateLimit.Value), 1))
		}
		s.reverseProxies[reverseProxyKey] = reverseProxy
	}

	return reverseProxy
}
