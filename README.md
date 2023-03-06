# Protty

----

<p align="center">
  <img height="250" alt="PROTTY" src="assets/logo/logo.png"/>
</p>

----

Protty is a HTTP proxy written in Go that redirects requests to a remote host.
The proxy intercepts and processes the requests before forwarding them to the desired host, and then sends the response back to the user.
In addition to proxying requests, Protty also offers features such as throttling requests to control the rate at which requests are sent, and intercepting and modifying the request and response body to add or modify data(in the nearest future release).
These capabilities make Protty a useful tool for a variety of purposes, such as testing applications, debugging network issues, or adding custom functionality to HTTP traffic.

## Usage

The following command will start a proxy on port 8080, and after starting, all traffic from port 8080 will be redirected to a remote host located at https://example.com

```shell
docker run -p8080:80 -e REMOTE_URI=https://example.com:443 mgerasimchuk/protty:v0.1.0
```

## Running options and runtime configuration

```
» ~  docker run -p8080:80 -it mgerasimchuk/protty:v0.1.0 /bin/sh -c 'protty start --help'  
Start the proxy

Usage:
  protty start [flags]

Examples:
  protty start --remote-uri https://www.githubstatus.com --throttle-rate-limit 2

Flags:
      --log-level string            On which host, the throttle rate limit should be applied | Env variable alias: LOG_LEVEL | Request header alias: X-PROTTY-LOG-LEVEL (default "debug")
      --local-port int              Verbosity level (panic, fatal, error, warn, info, debug, trace) | Env variable alias: LOCAL_PORT | Request header alias: X-PROTTY-LOCAL-PORT (default 80)
      --remote-uri string           Listening port for the proxy | Env variable alias: REMOTE_URI | Request header alias: X-PROTTY-REMOTE-URI (default "https://example.com:443")
      --throttle-rate-limit float   URI of the remote resource | Env variable alias: THROTTLE_RATE_LIMIT | Request header alias: X-PROTTY-THROTTLE-RATE-LIMIT
      --throttle-host string        How many requests can be send to the remote resource per second | Env variable alias: THROTTLE_HOST | Request header alias: X-PROTTY-THROTTLE-HOST
  -h, --help                        help for start

*Use CLI flags, environment variables or request headers to configure settings. The settings will be applied in the following priority: environment variables -> CLI flags -> request headers
```

## Dependencies

- SED implementation - https://github.com/rwtodd/Go.Sed/tree/55464686f9ef25a9d147120ab1f7f489eae471fd
- JQ implementation - https://github.com/itchyny/gojq/tree/v0.12.11
