package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"reflect"
	"strconv"
)

type Config struct {
	LocalPort         Option[int]
	RemoteURI         Option[string]
	ThrottleRateLimit Option[float64]
	ThrottleHost      Option[string]
	LogLevel          Option[string]
}

func GetConfig() *Config {
	cfg := Config{}
	e := reflect.ValueOf(cfg)
	for i := 0; i < e.NumField(); i++ {
		curOptionName := e.Type().Field(i).Name
		opt := reflect.ValueOf(&cfg).Elem().FieldByName(curOptionName).Addr()
		optName := opt.Elem().FieldByName("Name")

		// Set name in kebab case for all config fields
		optName.SetString(curOptionName)
	}

	// Set default
	cfg.LocalPort.Value = 80
	cfg.RemoteURI.Value = "https://example.com:443"
	cfg.LogLevel.Value = "debug"

	return &cfg
}

func (c *Config) ReadEnv() error {
	e := reflect.ValueOf(*c)
	for i := 0; i < e.NumField(); i++ {
		curOptionName := e.Type().Field(i).Name
		opt := reflect.ValueOf(c).Elem().FieldByName(curOptionName).Addr()
		optValue := opt.Elem().FieldByName("Value")

		// Lookup for the env variable (can be move to the separate function)
		envName := opt.MethodByName("GetEnvName").Call([]reflect.Value{})[0].String()
		if val, ok := os.LookupEnv(envName); ok {
			switch optValue.Kind() {
			case reflect.String:
				optValue.SetString(val)
			case reflect.Int:
				v, err := strconv.Atoi(val)
				if err != nil {
					return fmt.Errorf("parsing env variable %s: %w", envName, err)
				}
				optValue.SetInt(int64(v))
			case reflect.Float64:
				v, err := strconv.ParseFloat(val, 32)
				if err != nil {
					return fmt.Errorf("parsing env variable %s: %w", envName, err)
				}
				optValue.SetFloat(v)
			}
		}
	}
	return nil
}

func (c *Config) Validate() error {
	if _, err := logrus.ParseLevel(c.LogLevel.Value); err != nil {
		return fmt.Errorf("parse logrus LogLevel: %w", err)
	}
	if _, err := url.Parse(c.RemoteURI.Value); err != nil {
		return fmt.Errorf("parse RemoteURI: %w", err)
	}
	return nil
}

func (c *Config) GetLogLevelLogrus() logrus.Level {
	logLevel, _ := logrus.ParseLevel(c.LogLevel.Value)
	return logLevel
}
