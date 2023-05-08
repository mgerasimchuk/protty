package config

import (
	"crypto/md5"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"protty/pkg/util"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type StartCommandConfig struct {
	LogLevel                  Option[string]   `default:"debug" description:"On which host, the throttle rate limit should be applied"`
	LocalPort                 Option[int]      `default:"80" description:"Verbosity level (panic, fatal, error, warn, info, debug, trace)"`
	RemoteURI                 Option[string]   `default:"https://example.com:443" description:"Listening port for the proxy"`
	ThrottleRateLimit         Option[float64]  `description:"How many requests can be send to the remote resource per second"`
	TransformRequestUrlSED    Option[string]   `description:"SED expression for request URL transformation"`
	AdditionalRequestHeaders  Option[[]string] `description:"Array of additional request headers in format Header: Value"`
	TransformRequestBodySED   Option[[]string] `description:"Pipeline of SED expressions for request body transformation"`
	TransformRequestBodyJQ    Option[[]string] `description:"Pipeline of JQ expressions for request body transformation"`
	AdditionalResponseHeaders Option[[]string] `description:"Array of additional response headers in format Header: Value"`
	TransformResponseBodySED  Option[[]string] `description:"Pipeline of SED expressions for response body transformation"`
	TransformResponseBodyJQ   Option[[]string] `description:"Pipeline of JQ expressions for response body transformation"`
}

func GetStartCommandConfig() *StartCommandConfig {
	cfg := StartCommandConfig{}
	e := reflect.ValueOf(cfg)
	for i := 0; i < e.NumField(); i++ {
		opt := e.Type().Field(i)
		optAddr := reflect.ValueOf(&cfg).Elem().FieldByName(opt.Name).Addr()
		optValueField := optAddr.Elem().FieldByName("Value")

		// Set name in kebab case for all config fields
		optNameField := optAddr.Elem().FieldByName("Name")
		optNameField.SetString(opt.Name)

		// Set description from struct tag
		optDescriptionField := optAddr.Elem().FieldByName("Description")
		if tagValue := opt.Tag.Get("description"); tagValue != "" {
			optDescriptionField.SetString(tagValue)
		}

		// Set default from struct tag
		if tagValue := opt.Tag.Get("default"); tagValue != "" {
			if err := setOptValue(&optValueField, tagValue); err != nil {
				panic(fmt.Sprintf("parsing default tag of option %s: %s", opt.Name, err))
			}
		}
	}
	return &cfg
}

func (c *StartCommandConfig) SetFromEnv() error {
	e := reflect.ValueOf(*c)
	for i := 0; i < e.NumField(); i++ {
		opt := e.Type().Field(i)
		optAddr := reflect.ValueOf(c).Elem().FieldByName(opt.Name).Addr()
		optValueField := optAddr.Elem().FieldByName("Value")

		// Lookup for the env variable (can be move to the separate function)
		envName := optAddr.MethodByName("GetEnvName").Call([]reflect.Value{})[0].String()

		if optValueField.Kind() == reflect.Slice {
			envValuesSlice := []string{}
			for _, envPair := range os.Environ() {
				envPairSlice := strings.Split(envPair, "=")
				eName, eVal := envPairSlice[0], envPairSlice[1]
				if isMatch, _ := regexp.MatchString(envName+`(_\d+)?$`, eName); isMatch {
					envValuesSlice = append(envValuesSlice, eVal)
				}
			}
			if len(envValuesSlice) > 0 {
				if err := setOptValue(&optValueField, envValuesSlice); err != nil {
					return fmt.Errorf("%s envName - %s: %w", util.GetFuncName(setOptValue), envName, err)
				}
			}
		} else if val, ok := os.LookupEnv(envName); ok {
			if err := setOptValue(&optValueField, val); err != nil {
				return fmt.Errorf("%s envName - %s: %w", util.GetFuncName(setOptValue), envName, err)
			}
		}
	}
	return nil
}

// TODO add availableInRuntime mapstructure flag and based on this flag throw the error if the user try to change cfg for this field through the http headers
func (c *StartCommandConfig) SetFromHTTPRequestHeaders(header http.Header, logger *logrus.Logger) error {
	e := reflect.ValueOf(*c)
	for i := 0; i < e.NumField(); i++ {
		opt := e.Type().Field(i)
		optAddr := reflect.ValueOf(c).Elem().FieldByName(opt.Name).Addr()
		optValueField := optAddr.Elem().FieldByName("Value")

		headerName := optAddr.MethodByName("GetHeaderName").Call([]reflect.Value{})[0].String()
		if values := header.Values(headerName); len(values) > 0 {
			if err := setOptValueFromHTTPRequestHeader(&optValueField, values); err != nil {
				return fmt.Errorf("%s headerName - %s: %w", util.GetFuncName(setOptValueFromHTTPRequestHeader), headerName, err)
			}
			if logger != nil {
				logger.Debugf("%s config value has been changed to `%v' based on %s request header", opt.Name, values, headerName)
			}
		}
	}
	return nil
}

func (c *StartCommandConfig) Validate() error {
	if _, err := logrus.ParseLevel(c.LogLevel.Value); err != nil {
		return fmt.Errorf("%s: %w", util.GetFuncName(logrus.ParseLevel), err)
	}
	if _, err := url.Parse(c.RemoteURI.Value); err != nil {
		return fmt.Errorf("%s: %w", util.GetFuncName(url.Parse), err)
	}
	e := reflect.ValueOf(*c)
	for i := 0; i < e.NumField(); i++ {
		opt := e.Type().Field(i)
		optAddr := reflect.ValueOf(c).Elem().FieldByName(opt.Name).Addr()
		optValueField := optAddr.Elem().FieldByName("IsAddedToCLI")
		if !optValueField.Bool() {
			return fmt.Errorf("configuration field '%s' has not been added to the CLI flags", opt.Name)
		}
	}
	return nil
}

func (c *StartCommandConfig) GetLogLevelLogrus() logrus.Level {
	logLevel, _ := logrus.ParseLevel(c.LogLevel.Value)
	return logLevel
}

func (c *StartCommandConfig) GetStateHash() string {
	fieldsDump := ""
	e := reflect.ValueOf(*c)
	for i := 0; i < e.NumField(); i++ {
		opt := e.Type().Field(i)
		optAddr := reflect.ValueOf(c).Elem().FieldByName(opt.Name).Addr()
		optValueField := optAddr.Elem().FieldByName("Value")
		fieldsDump += fmt.Sprintf("%v", optValueField)
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(fieldsDump)))
}

// TODO refactor "val any" to generic should accept string and []string
func setOptValue(optValue *reflect.Value, val any) error {
	switch optValue.Kind() {
	case reflect.String:
		optValue.SetString(val.(string))
	case reflect.Int:
		v, err := strconv.Atoi(val.(string))
		if err != nil {
			return fmt.Errorf("%s: %w", util.GetFuncName(strconv.Atoi), err)
		}
		optValue.SetInt(int64(v))
	case reflect.Float64:
		v, err := strconv.ParseFloat(val.(string), 32)
		if err != nil {
			return fmt.Errorf("%s: %w", util.GetFuncName(strconv.ParseFloat), err)
		}
		optValue.SetFloat(v)
	case reflect.Slice:
		valSlice := val.([]string)
		optValue.Set(reflect.MakeSlice(optValue.Type(), len(valSlice), len(valSlice)))
		for i := 0; i < len(valSlice); i++ {
			optValue.Index(i).SetString(valSlice[i])
		}
	}
	return nil
}

func setOptValueFromHTTPRequestHeader(optValue *reflect.Value, val []string) error {
	if optValue.Kind() == reflect.Slice {
		return setOptValue(optValue, val)
	}
	return setOptValue(optValue, val[0])
}
