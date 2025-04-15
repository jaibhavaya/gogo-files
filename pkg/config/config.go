package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL          string `env:"DATABASE_URL" required:"true"`
	QueueURL             string `env:"QUEUE_URL" required:"true"`
	AWSRegion            string `env:"AWS_REGION" default:"us-west-1"`
	S3Bucket             string `env:"S3_BUCKET" required:"true"`
	S3Endpoint           string `env:"S3_ENDPOINT"`
	Environment          string `env:"ENVIRONMENT" default:"development"`
	EncryptionKey        string `env:"ENCRYPTION_KEY" default:"default-dev-key-please-change-in-production"`
	OnedriveClientID     string `env:"ONEDRIVE_CLIENT_ID" default:"your-client-id"`
	OnedriveClientSecret string `env:"ONEDRIVE_CLIENT_SECRET" default:"your-client-secret"`
}

func FromEnv() (*Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	_ = v.ReadInConfig()

	config := &Config{}
	t := reflect.TypeOf(*config)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		// Set default if specified
		if defaultVal := field.Tag.Get("default"); defaultVal != "" {
			v.SetDefault(envTag, defaultVal)
		}

		// Check if required
		if field.Tag.Get("required") == "true" && !v.IsSet(envTag) {
			return nil, fmt.Errorf("%s is required but not set", envTag)
		}

		// Set field value
		reflect.ValueOf(config).Elem().Field(i).SetString(v.GetString(envTag))
	}

	return config, nil
}
