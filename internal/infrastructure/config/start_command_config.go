package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"reflect"
	"strconv"
)

type StartCommandConfig struct {
	LocalPort         Option[int]     `default:"80" description:"Verbosity level (panic, fatal, error, warn, info, debug, trace)"`
	RemoteURI         Option[string]  `default:"https://example.com:443" description:"Listening port for the proxy"`
	ThrottleRateLimit Option[float64] `description:"URI of the remote resource"`
	ThrottleHost      Option[string]  `description:"How many requests can be send to the remote resource per second"`
	LogLevel          Option[string]  `default:"debug" description:"On which host, the throttle rate limit should be applied"`
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

func (c *StartCommandConfig) ReadEnv() error {
	e := reflect.ValueOf(*c)
	for i := 0; i < e.NumField(); i++ {
		opt := e.Type().Field(i)
		optAddr := reflect.ValueOf(c).Elem().FieldByName(opt.Name).Addr()
		optValueField := optAddr.Elem().FieldByName("Value")

		// Lookup for the env variable (can be move to the separate function)
		envName := optAddr.MethodByName("GetEnvName").Call([]reflect.Value{})[0].String()
		if val, ok := os.LookupEnv(envName); ok {
			if err := setOptValue(&optValueField, val); err != nil {
				return fmt.Errorf("parsing env variable %s: %w", envName, err)
			}
		}
	}
	return nil
}

func (c *StartCommandConfig) Validate() error {
	if _, err := logrus.ParseLevel(c.LogLevel.Value); err != nil {
		return fmt.Errorf("parse logrus LogLevel: %w", err)
	}
	if _, err := url.Parse(c.RemoteURI.Value); err != nil {
		return fmt.Errorf("parse RemoteURI: %w", err)
	}
	return nil
}

func (c *StartCommandConfig) GetLogLevelLogrus() logrus.Level {
	logLevel, _ := logrus.ParseLevel(c.LogLevel.Value)
	return logLevel
}

func setOptValue(optValue *reflect.Value, val string) error {
	switch optValue.Kind() {
	case reflect.String:
		optValue.SetString(val)
	case reflect.Int:
		v, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("strconv.Atoi: %w", err)
		}
		optValue.SetInt(int64(v))
	case reflect.Float64:
		v, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return fmt.Errorf("strconv.ParseFloat: %w", err)
		}
		optValue.SetFloat(v)
	}
	return nil
}
