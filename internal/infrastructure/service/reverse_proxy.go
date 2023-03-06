package service

import (
	"bytes"
	"fmt"
	"github.com/graze/go-throttled"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"protty/internal/infrastructure/config"
	"protty/pkg/util"
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

func (s *ReverseProxyService) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	s.logRequestPayload(req)
	s.serveReverseProxy(res, req)
}

func (s *ReverseProxyService) logRequestPayload(req *http.Request) {
	s.logger.WithField("method", req.Method).WithField("path", req.URL.Path).Infof("Send request to %s", req.URL.Host)
	// TODO add tracing log with body and other params like headers
}

// Serve a reverse proxy for a given url
func (s *ReverseProxyService) serveReverseProxy(res http.ResponseWriter, req *http.Request) {
	cfg := s.getOverrideConfig(req)

	host := strings.ReplaceAll(strings.ReplaceAll(cfg.RemoteURI.Value, "https://", ""), "http://", "")
	req.URL.Host = host
	req.Host = host

	reverseProxy := s.getReverseProxyByParams(*cfg)

	// Deleting encoding to keep availability for changing response
	req.Header.Del("Accept-Encoding")

	reverseProxy.ServeHTTP(res, req)
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
		sourceResponse, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(ioutil.ReadAll), err)
			return nil
		}
		resp.Body.Close()
		modifiedResponse := sourceResponse

		for _, sedExpr := range cfg.TransformResponseBodySED.Value {
			modifiedResponse, sourceResponse, err = util.SED(sedExpr, modifiedResponse)
			if err != nil {
				s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(util.SED), err)
				return nil
			}
			s.logger.Debugf("ModifyResponse: %s", getChangesLogMessage(sourceResponse, modifiedResponse, sedExpr, cfg.TransformResponseBodySED))
		}
		for _, jqExpr := range cfg.TransformResponseBodyJQ.Value {
			modifiedResponse, sourceResponse, err = util.JQ(jqExpr, modifiedResponse)
			if err != nil {
				s.logger.Errorf("%s: %s: %s", util.GetCurrentFuncName(), util.GetFuncName(util.JQ), err)
				return nil
			}
			s.logger.Debugf("ModifyResponse: %s", getChangesLogMessage(sourceResponse, modifiedResponse, jqExpr, cfg.TransformResponseBodyJQ))
		}

		buf := bytes.NewBufferString("")
		buf.Write(modifiedResponse)
		resp.Body = ioutil.NopCloser(buf)

		resp.Body = ioutil.NopCloser(bytes.NewBuffer(modifiedResponse))
		resp.Header["Content-Length"] = []string{strconv.Itoa(len(modifiedResponse))}
		resp.ContentLength = int64(len(modifiedResponse))
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
		return fmt.Sprintf("the '%s' %s expression didn't change the response", expr, o.Name)
	} else {
		// TODO add diff to log
		return fmt.Sprintf("the '%s' %s expression changed the response. Length difference: %d",
			expr, o.Name, int(math.Abs(float64(len(source)-len(modified)))))
	}
}
