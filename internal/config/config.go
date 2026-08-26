package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	yamlv2 "gopkg.in/yaml.v2"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/chubin/wttr.in/internal/cache"
	"github.com/chubin/wttr.in/internal/ip"
	"github.com/chubin/wttr.in/internal/location"
	"github.com/chubin/wttr.in/internal/logging"
	"github.com/chubin/wttr.in/internal/renderer"
	"github.com/chubin/wttr.in/internal/server"
	"github.com/chubin/wttr.in/internal/uplink"
	"github.com/chubin/wttr.in/internal/weather"
)

// Config is the root configuration structure for the entire service.
type Config struct {
	// Geo contains location-related settings (IP geolocation, default city, etc.)
	Geo *location.Config `yaml:"geo"`

	// IP contains settings for IP address parsing and geolocation lookup
	IP *ip.Config `yaml:"ip"`

	// Weather contains weather data source configuration (WWO, etc.)
	Weather *weather.Config `yaml:"weather"`

	// Cache defines caching behavior for weather data, location results, etc.
	Cache *cache.Config `yaml:"cache"`

	// Logging controls log level, format, output destinations, and tracing
	Logging *logging.Config `yaml:"logging"`

	// Uplink contains configuration for upstream modules
	Uplink *uplink.Config `yaml:"uplink"`

	// Server defines HTTP server settings (port, timeouts, TLS, etc.)
	Server *server.Config `yaml:"server"`

	// Renderer configures how responses are rendered (subprocess renderers, templates, etc.)
	Renderer *renderer.Config `yaml:"renderer"`
}

var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?\}`)

// LoadFromYAML loads configuration from a YAML file and returns a pointer to Config
func LoadFromYAML(filePath string) (*Config, error) {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	data, err = expandEnv(data)
	if err != nil {
		return nil, fmt.Errorf("error expanding environment variables in YAML: %v", err)
	}

	// Create a new Config instance
	config := &Config{}

	// Unmarshal YAML data into the Config struct with strict checking for unknown fields
	err = yamlv2.UnmarshalStrict(data, config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %v", err)
	}

	return config, nil
}

// expandEnv substitutes ${NAME} and ${NAME:-default} placeholders only inside
// already-parsed YAML scalar nodes. Environment values are never interpreted as
// YAML source or shell syntax, so newlines, quotes, and ${...} inside a value
// remain data inside that scalar.
func expandEnv(data []byte) ([]byte, error) {
	var doc yamlv3.Node
	if err := yamlv3.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	expandEnvNode(&doc)

	return yamlv3.Marshal(&doc)
}

func expandEnvNode(node *yamlv3.Node) {
	if node.Kind == yamlv3.ScalarNode {
		expanded, changed := expandEnvValue(node.Value)
		if changed {
			envOnly := node.Value != "" && envVarPattern.ReplaceAllString(node.Value, "") == ""
			node.Value = expanded
			if envOnly && node.Style == 0 {
				node.Tag = yamlTag(expanded)
			} else {
				node.Tag = "!!str"
			}
		}
		return
	}

	for _, child := range node.Content {
		expandEnvNode(child)
	}
}

func expandEnvValue(value string) (string, bool) {
	changed := false
	expanded := envVarPattern.ReplaceAllStringFunc(value, func(match string) string {
		changed = true
		groups := envVarPattern.FindStringSubmatch(match)
		if envValue, ok := os.LookupEnv(groups[1]); ok {
			return envValue
		}
		if len(groups) == 4 {
			return groups[3]
		}
		return ""
	})

	return expanded, changed
}

func yamlTag(value string) string {
	switch {
	case value == "true" || value == "false":
		return "!!bool"
	case isYAMLInt(value):
		return "!!int"
	default:
		return "!!str"
	}
}

func isYAMLInt(value string) bool {
	value = strings.TrimPrefix(strings.TrimPrefix(value, "-"), "+")
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (c *Config) MarshalYAML() ([]byte, error) {
	return yamlv2.Marshal(c)
}
