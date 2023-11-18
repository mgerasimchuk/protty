# Protty

----

<p align="center">
  <img height="250" alt="PROTTY" src="https://github.com/mgerasimchuk/protty/raw/master/assets/logo/logo.png"/>
</p>

----

[![Release](https://img.shields.io/github/v/release/mgerasimchuk/protty)](https://github.com/mgerasimchuk/protty/releases)
[![Docker pulls](https://img.shields.io/docker/pulls/mgerasimchuk/protty)](https://hub.docker.com/r/mgerasimchuk/protty)\
[![Lint Golangci](https://github.com/mgerasimchuk/protty/actions/workflows/lint-golangci.yml/badge.svg)](https://github.com/mgerasimchuk/protty/actions/workflows/lint-golangci.yml)\
[![Test (unit)](https://github.com/mgerasimchuk/protty/actions/workflows/test-unit.yml/badge.svg)](https://github.com/mgerasimchuk/protty/actions/workflows/test-unit.yml)
[![Coverage (unit)](https://github.com/mgerasimchuk/protty/wiki/assets/coverage/unit/coverage.svg)](https://github.com/mgerasimchuk/protty/wiki/Test-coverage-report)\
[![Test (integration)](https://github.com/mgerasimchuk/protty/actions/workflows/test-integration.yml/badge.svg)](https://github.com/mgerasimchuk/protty/actions/workflows/test-integration.yml)
[![Coverage (integration)](https://github.com/mgerasimchuk/protty/wiki/assets/coverage/integration/coverage.svg)](https://github.com/mgerasimchuk/protty/wiki/Test-coverage-report)

Protty is an HTTP proxy written in Go that redirects, intercepts, and modifies both requests to a remote host and their
corresponding responses.
The proxy intercepts and processes the requests before forwarding them to the desired host, and then changed the
responses and sends back to the user.
In addition to proxying requests, Protty also offers features such as throttling requests to control the rate at which
requests are sent.
These capabilities make Protty a useful tool for a variety of purposes, such as testing applications, debugging network issues, or adding custom functionality to HTTP traffic.

## Usage

The following command will start a proxy on port 8080, and after starting, all traffic from port 8080 will be redirected to a remote host located at https://example.com

```shell
docker run -p8080:80 -e REMOTE_URI=https://example.com:443 mgerasimchuk/protty:v0.4.6
```

## Supported Backends

- Docker - https://hub.docker.com/r/mgerasimchuk/protty

## Running options and runtime configuration

```
» ~  docker run -p8080:80 -it mgerasimchuk/protty:v0.4.6 /bin/sh -c 'protty start --help'  
Start the proxy

Usage:
  protty start [flags]

Examples:
  # Start the proxy with default values
  protty start
  
  # Start the proxy with specific log level
  protty start --log-level info

  # Start the proxy with a specific local port
  protty start --local-port 8080
  
  # Start the proxy with a specific remote URI and specific throttle rate limit 
  protty start --remote-uri https://www.githubstatus.com --throttle-rate-limit 2

  # Start the proxy with a specific additional request headers
  protty start --additional-request-headers 'Authorization: Bearer authtoken-with:any:symbols' --additional-request-headers 'X-Another-One: another-value'

  # Start the proxy with a specific SED expression for response transformation
  protty start --transform-response-body-sed 's|old|new|g'

  # Start the proxy with a specific SED expressions pipeline for response transformation
  protty start --transform-response-body-sed 's|old|new-stage-1|g' --transform-response-body-sed 's|new-stage-1|new-stage-2|g'

  # Start the proxy with a specific SED expressions pipeline for response transformation (configured with env)
  TRANSFORM_RESPONSE_BODY_SED_0='s|old|new-stage-1|g' TRANSFORM_RESPONSE_BODY_SED_1='s|new-stage-1|new-stage-2|g' protty start

  # Start the proxy with a specific JQ expressions pipeline for response transformation
  protty start --transform-response-body-jq '.[] | .id'

Flags:
      --log-level string                          Verbosity level (panic, fatal, error, warn, info, debug, trace) | Env variable alias: LOG_LEVEL | Request header alias: X-PROTTY-LOG-LEVEL (default "debug")
      --local-port int                            Listening port for the proxy | Env variable alias: LOCAL_PORT | Request header alias: X-PROTTY-LOCAL-PORT (default 80)
      --remote-uri string                         URI of the remote resource | Env variable alias: REMOTE_URI | Request header alias: X-PROTTY-REMOTE-URI (default "https://example.com:443")
      --throttle-rate-limit float                 How many requests can be send to the remote resource per second | Env variable alias: THROTTLE_RATE_LIMIT | Request header alias: X-PROTTY-THROTTLE-RATE-LIMIT
      --transform-request-url-sed string          SED expression for request URL transformation | Env variable alias: TRANSFORM_REQUEST_URL_SED | Request header alias: X-PROTTY-TRANSFORM-REQUEST-URL-SED
      --additional-request-headers stringArray    Array of additional request headers in format Header: Value | Env variable alias: ADDITIONAL_REQUEST_HEADERS | Request header alias: X-PROTTY-ADDITIONAL-REQUEST-HEADERS
      --transform-request-body-sed stringArray    Pipeline of SED expressions for request body transformation | Env variable alias: TRANSFORM_REQUEST_BODY_SED | Request header alias: X-PROTTY-TRANSFORM-REQUEST-BODY-SED
      --transform-request-body-jq stringArray     Pipeline of JQ expressions for request body transformation | Env variable alias: TRANSFORM_REQUEST_BODY_JQ | Request header alias: X-PROTTY-TRANSFORM-REQUEST-BODY-JQ
      --additional-response-headers stringArray   Array of additional response headers in format Header: Value | Env variable alias: ADDITIONAL_RESPONSE_HEADERS | Request header alias: X-PROTTY-ADDITIONAL-RESPONSE-HEADERS
      --transform-response-body-sed stringArray   Pipeline of SED expressions for response body transformation | Env variable alias: TRANSFORM_RESPONSE_BODY_SED | Request header alias: X-PROTTY-TRANSFORM-RESPONSE-BODY-SED
      --transform-response-body-jq stringArray    Pipeline of JQ expressions for response body transformation | Env variable alias: TRANSFORM_RESPONSE_BODY_JQ | Request header alias: X-PROTTY-TRANSFORM-RESPONSE-BODY-JQ
  -h, --help                                      help for start

*Use CLI flags, environment variables or request headers to configure settings. The settings will be applied in the following priority: environment variables -> CLI flags -> request headers
```

## Dependencies

- SED implementation - https://github.com/rwtodd/Go.Sed/tree/ba3e9c1
- JQ implementation - https://github.com/itchyny/gojq/tree/v0.12.11
