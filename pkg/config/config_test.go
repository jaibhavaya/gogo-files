package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromEnv(t *testing.T) {
	// All environment variables we'll be manipulating
	allVars := []string{
		"DATABASE_URL", "QUEUE_URL", "AWS_REGION", "S3_BUCKET",
		"S3_ENDPOINT", "ENVIRONMENT", "ENCRYPTION_KEY",
		"ONEDRIVE_CLIENT_ID", "ONEDRIVE_CLIENT_SECRET",
	}

	// Test cases
	tests := []struct {
		name     string
		envVars  map[string]string
		envFile  string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name:    "missing_required_fields",
			envVars: map[string]string{},
			wantErr: true,
		},
		{
			name: "default_values",
			envVars: map[string]string{
				"DATABASE_URL": "test-db-url",
				"QUEUE_URL":    "test-queue-url",
				"S3_BUCKET":    "test-bucket",
			},
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, "us-west-1", c.AWSRegion)
				assert.Equal(t, "development", c.Environment)
				assert.Equal(t, "default-dev-key-please-change-in-production", c.EncryptionKey)
			},
		},
		{
			name: "env_overrides_defaults",
			envVars: map[string]string{
				"DATABASE_URL":           "custom-db-url",
				"QUEUE_URL":              "custom-queue-url",
				"AWS_REGION":             "eu-central-1",
				"S3_BUCKET":              "custom-bucket",
				"S3_ENDPOINT":            "custom-endpoint",
				"ENVIRONMENT":            "production",
				"ENCRYPTION_KEY":         "secure-key",
				"ONEDRIVE_CLIENT_ID":     "real-client-id",
				"ONEDRIVE_CLIENT_SECRET": "real-client-secret",
			},
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, "custom-db-url", c.DatabaseURL)
				assert.Equal(t, "eu-central-1", c.AWSRegion)
				assert.Equal(t, "production", c.Environment)
			},
		},
		{
			name: "dot_env_file",
			envFile: `DATABASE_URL=env-file-db-url
QUEUE_URL=env-file-queue-url
S3_BUCKET=env-file-bucket
AWS_REGION=ap-southeast-1`,
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, "env-file-db-url", c.DatabaseURL)
				assert.Equal(t, "ap-southeast-1", c.AWSRegion)
			},
		},
		{
			name: "env_overrides_dot_env",
			envVars: map[string]string{
				"DATABASE_URL": "env-var-db-url",
				"S3_ENDPOINT":  "env-var-endpoint",
			},
			envFile: `DATABASE_URL=env-file-db-url
QUEUE_URL=env-file-queue-url
S3_BUCKET=env-file-bucket`,
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, "env-var-db-url", c.DatabaseURL)
				assert.Equal(t, "env-file-queue-url", c.QueueURL)
				assert.Equal(t, "env-var-endpoint", c.S3Endpoint)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup: backup and clear environment
			envBackup := backupEnv(allVars)
			defer restoreEnv(envBackup)

			// Apply test environment variables
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}

			// Create .env file if specified
			if tc.envFile != "" {
				require.NoError(t, os.WriteFile(".env", []byte(tc.envFile), 0644))
				defer os.Remove(".env")
			} else {
				os.Remove(".env") // Ensure no file exists
			}

			// Run test
			config, err := FromEnv()

			// Assertions
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)

			if tc.validate != nil {
				tc.validate(t, config)
			}
		})
	}
}

// Helper functions to make the test more readable
func backupEnv(keys []string) map[string]string {
	backup := make(map[string]string)
	for _, key := range keys {
		if val, exists := os.LookupEnv(key); exists {
			backup[key] = val
		}
		os.Unsetenv(key)
	}
	return backup
}

func restoreEnv(backup map[string]string) {
	for k, v := range backup {
		os.Setenv(k, v)
	}
}
