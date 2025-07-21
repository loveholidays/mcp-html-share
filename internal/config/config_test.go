package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) TestValidateWithValidConfig() {
	config := Config{
		BucketName: "test-bucket",
		Transport:  "stdio",
		HTTPPort:   "8080",
		HealthPort: "9090",
	}

	err := config.Validate()
	s.NoError(err)
}

func (s *ConfigTestSuite) TestValidateWithMissingBucketName() {
	config := Config{
		Transport:  "stdio",
		HTTPPort:   "8080",
		HealthPort: "9090",
	}

	err := config.Validate()
	s.Error(err)
	s.Contains(err.Error(), "bucket flag is required")
}

func (s *ConfigTestSuite) TestValidateWithInvalidTransport() {
	config := Config{
		BucketName: "test-bucket",
		Transport:  "invalid",
		HTTPPort:   "8080",
		HealthPort: "9090",
	}

	err := config.Validate()
	s.Error(err)
	s.Contains(err.Error(), "transport must be either 'stdio' or 'http'")
}

func (s *ConfigTestSuite) TestValidateWithValidHTTPTransport() {
	config := Config{
		BucketName: "test-bucket",
		Transport:  "http",
		HTTPPort:   "8080",
		HealthPort: "9090",
	}

	err := config.Validate()
	s.NoError(err)
}
