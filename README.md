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
docker run -p8080:80 -e REMOTE_URI=https://example.com:443 mgerasimchuk/protty:v0.0.1
```

## Custom configuration

Define the following settings to configure protty.

| Option            | Description                                                     | Env variable        | Flag                | Request Header               | Optional | Default             |
|-------------------|-----------------------------------------------------------------|---------------------|---------------------|------------------------------|:--------:|---------------------|
| LocalPort         | Verbosity level (panic, fatal, error, warn, info, debug, trace) | LOCAL_PORT          | local-port          | X-PROTTY-LOCAL-PORT          |   yes    | 80                  |
| RemoteURI         | Listening port for the proxy                                    | REMOTE_URI          | remote-uri          | X-PROTTY-REMOTE-URI          |   yes    | https://example.com |
| ThrottleRateLimit | URI of the remote resource                                      | THROTTLE_RATE_LIMIT | throttle-rate-limit | X-PROTTY-THROTTLE-RATE-LIMIT |   yes    |                     |
| ThrottleHost      | How many requests can be send to the remote resource per second | THROTTLE_HOST       | throttle-host       | X-PROTTY-THROTTLE-HOST       |   yes    |                     |
| LogLevel          | On which host, the throttle rate limit should be applied        | LOG_LEVEL           | log-level           | X-PROTTY-LOG-LEVEL           |   yes    | debug               |

Use environment variables or application flags or request headers to configure the app.

The settings will be applied in the following priority: environment variables -> command flags -> request headers

You can also get the full list of available commands and flags by using the help:

```
» ~  docker run -it mgerasimchuk/protty:v0.0.1 /bin/sh -c '/protty start --help'  
Start the proxy

Usage:
  protty start [flags]

Flags:
      --log-level string            Verbosity level (panic, fatal, error, warn, info, debug, trace) (default "debug")
      --local-port int              Listening port for the proxy (default 80)
      --remote-uri string           URI of the remote resource (default "https://example.com:443")
      --throttle-rate-limit float   How many requests can be send to the remote resource per second
      --throttle-host string        On which host, the throttle rate limit should be applied
  -h, --help                        help for start

*Use environment variables (for example, REMOTE_URI) or request headers (for example, X-PROTTY-REMOTE-URI) to configure settings. The settings will be applied in the following priority: environment variables -> command flags -> request headers
```
