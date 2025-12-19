package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/graze/go-throttled"
	"github.com/mgerasimchuk/protty/internal/infrastructure/config"
	"github.com/mgerasimchuk/protty/pkg/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type ReverseProxyService struct {
	srv            *http.Server
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

	s.logger.Infof("Start listen proxy on :%d port with config: %+v", s.cfg.LocalPort.Value, s.cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequestAndRedirect)
	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.LocalPort.Value),
		Handler: mux,
	}
	return s.srv.ListenAndServe()
}

func (s *ReverseProxyService) Stop(ctx context.Context) error {
	s.logger.Infof("Stoping proxy")
	return s.srv.Shutdown(ctx)
}

func (s *ReverseProxyService) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	s.serveReverseProxy(res, req)
	s.logRequestPayload(req)
}

func (s *ReverseProxyService) logRequestPayload(req *http.Request) {
	s.logger.WithField("method", req.Method).WithField("path", req.URL.Path).Infof("Request have been sent to %s", req.URL.Host)
	// TODO add tracing log with body and other params like headers
}

// Serve a reverse proxy for a given url
func (s *ReverseProxyService) serveReverseProxy(res http.ResponseWriter, req *http.Request) {
	cfg := s.getOverrideConfig(req)
	reverseProxy := s.getReverseProxyByParams(*cfg)
	modifiedReq := s.getModifiedRequest(*cfg, req)

	reverseProxy.ServeHTTP(res, modifiedReq)
}

func (s *ReverseProxyService) getModifiedRequest(cfg config.StartCommandConfig, req *http.Request) *http.Request {
	modifiedReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
	if err != nil {
		s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(http.NewRequest), err)
		return req
	}

	for headerKey, headerValues := range req.Header {
		headerKey = strings.ToLower(headerKey)
		if headerKey == "accept-encoding" || // Skipping encoding to keep availability for changing response (for example in case of using gzip, we would not be able to make a replacing in response body)
			strings.HasPrefix(headerKey, "x-protty-") { // Skipping x-protty-* headers cos they need only for protty
			continue
		}
		for _, headerValue := range headerValues {
			modifiedReq.Header.Add(headerKey, headerValue)
		}
	}

	host := strings.ReplaceAll(strings.ReplaceAll(cfg.RemoteURI.Value, "https://", ""), "http://", "")
	modifiedReq.Host, modifiedReq.URL.Host = host, host

	// Transform request URL
	if cfg.TransformRequestUrlSED.Value != "" {
		if modifiedURLRaw, _, err := util.SED(cfg.TransformRequestUrlSED.Value, []byte(modifiedReq.URL.String())); err == nil {
			sourceURLRaw := modifiedReq.URL.String()
			if modifiedURL, err := url.Parse(strings.Trim(string(modifiedURLRaw), "\n")); err == nil { // TODO remove trim (currently it is a hotfix, cos the util.SED added \n at the end unexpectedly)
				modifiedReq.URL = modifiedURL
				s.logger.Debugf("ModifyRequestURL: %s", getChangesLogMessage([]byte(sourceURLRaw), modifiedURLRaw, cfg.TransformRequestUrlSED.Value, cfg.TransformRequestUrlSED))
			} else {
				s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(url.Parse), err)
				return req
			}
		} else {
			s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(util.SED), err)
		}
	}

	// Add request headers
	for _, header := range cfg.AdditionalRequestHeaders.Value {
		kv := strings.SplitN(header, ": ", 2)
		if len(kv) != 2 {
			s.logger.Errorf("%s: %s: %s - %+v", util.GetCurrentFuncName(), util.GetFuncName(strings.SplitN), "returns not 2 values", kv)
			continue
		}
		modifiedReq.Header.Add(kv[0], kv[1])
	}

	sourceRequestBody, err := io.ReadAll(modifiedReq.Body)
	if err != nil {
		s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(io.ReadAll), err)
		return req
	}
	if err = modifiedReq.Body.Close(); err != nil {
		s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(modifiedReq.Body.Close), err)
		return req
	}

	modifiedRequestBody := sourceRequestBody

	// Transform request body with SED
	for _, sedExpr := range cfg.TransformRequestBodySED.Value {
		modifiedRequestBody, sourceRequestBody, err = util.SED(sedExpr, modifiedRequestBody)
		if err != nil {
			s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(util.SED), err)
			return req
		}
		s.logger.Debugf("ModifyRequestBody: %s", getChangesLogMessage(sourceRequestBody, modifiedRequestBody, sedExpr, cfg.TransformRequestBodySED))
	}

	// Transform request body with JQ
	for _, jqExpr := range cfg.TransformRequestBodyJQ.Value {
		modifiedRequestBody, sourceRequestBody, err = util.JQ(jqExpr, modifiedRequestBody)
		if err != nil {
			s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(util.JQ), err)
			return req
		}
		s.logger.Debugf("ModifyRequestBody: %s", getChangesLogMessage(sourceRequestBody, modifiedRequestBody, jqExpr, cfg.TransformRequestBodySED))
	}

	modifiedReq.Body = io.NopCloser(bytes.NewBuffer(modifiedRequestBody))
	modifiedReq.ContentLength = int64(len(modifiedRequestBody))

	return modifiedReq
}

func (s *ReverseProxyService) getReverseProxyByParams(cfg config.StartCommandConfig) *httputil.ReverseProxy {
	var reverseProxy *httputil.ReverseProxy
	var ok bool

	// TODO memory leak - cos for every uniq combination, additional reverseProxy is creating (possible solution: make the LRU cache *with fixed size)
	if reverseProxy, ok = s.reverseProxies[cfg.GetStateHash()]; !ok {
		remoteURL, _ := url.Parse(cfg.RemoteURI.Value)
		reverseProxy = httputil.NewSingleHostReverseProxy(remoteURL)
		if cfg.ThrottleRateLimit.Value != 0 {
			// Another way to throttle requests on the handler side: https://github.com/go-chi/chi/blob/878319e482623b6e9c5787147e5b481f8879c49e/_examples/limits/main.go#L75
			reverseProxy.Transport = throttled.NewTransport(http.DefaultTransport, rate.NewLimiter(rate.Limit(cfg.ThrottleRateLimit.Value), 1))
		}
		reverseProxy.ModifyResponse = s.getModifyResponseFunc(cfg)
		s.reverseProxies[cfg.GetStateHash()] = reverseProxy
	}

	return reverseProxy
}

func (s *ReverseProxyService) getModifyResponseFunc(cfg config.StartCommandConfig) func(resp *http.Response) error {
	return func(resp *http.Response) error {
		sourceResponseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(io.ReadAll), err)
			return nil
		}
		if err = resp.Body.Close(); err != nil {
			s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(resp.Body.Close), err)
			return nil
		}

		modifiedResponseBody := sourceResponseBody

		// Transform response body with SED
		for _, sedExpr := range cfg.TransformResponseBodySED.Value {
			modifiedResponseBody, sourceResponseBody, err = util.SED(sedExpr, modifiedResponseBody)
			if err != nil {
				s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(util.SED), err)
				return nil
			}
			s.logger.Debugf("ModifyResponseBody: %s", getChangesLogMessage(sourceResponseBody, modifiedResponseBody, sedExpr, cfg.TransformResponseBodySED))
		}

		// Transform response body with SED
		for _, jqExpr := range cfg.TransformResponseBodyJQ.Value {
			modifiedResponseBody, sourceResponseBody, err = util.JQ(jqExpr, modifiedResponseBody)
			if err != nil {
				s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(util.JQ), err)
				return nil
			}
			s.logger.Debugf("ModifyResponseBody: %s", getChangesLogMessage(sourceResponseBody, modifiedResponseBody, jqExpr, cfg.TransformResponseBodyJQ))
		}

		buf := bytes.NewBufferString("")
		buf.Write(modifiedResponseBody)
		resp.Body = io.NopCloser(buf)

		resp.Body = io.NopCloser(bytes.NewBuffer(modifiedResponseBody))
		resp.Header["Content-Length"] = []string{strconv.Itoa(len(modifiedResponseBody))}
		resp.ContentLength = int64(len(modifiedResponseBody))

		// Add response headers
		for _, header := range cfg.AdditionalResponseHeaders.Value {
			kv := strings.SplitN(header, ": ", 2)
			if len(kv) != 2 {
				s.logger.Errorf("%s: %s: %s - %+v", util.GetCurrentFuncName(), util.GetFuncName(strings.SplitN), "returns not 2 values", kv)
				continue
			}
			resp.Header.Add(kv[0], kv[1])
		}

		return nil
	}
}

func (s *ReverseProxyService) getOverrideConfig(req *http.Request) *config.StartCommandConfig {
	cfg := *s.cfg
	if err := cfg.SetFromHTTPRequestHeaders(req.Header, s.logger); err != nil {
		s.logger.Errorf("%s: %s. Reverting to original config", util.GetFuncName(cfg.SetFromHTTPRequestHeaders), err)
		cfg = *s.cfg
	}
	if err := cfg.Validate(); err != nil {
		s.logger.Errorf("%s: %s. Reverting to original config", util.GetFuncName(cfg.Validate), err)
		cfg = *s.cfg
	}
	return &cfg
}

func getChangesLogMessage[T config.OptionValueType](source, modified []byte, expr string, o config.Option[T]) string {
	if string(source) == string(modified) {
		return fmt.Sprintf("the '%s' %s expression didn't change the data", expr, o.Name)
	} else {
		// TODO add diff to log
		return fmt.Sprintf("the '%s' %s expression changed the data. Length difference: %d",
			expr, o.Name, int(math.Abs(float64(len(source)-len(modified)))))
	}
}
