package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFromYAMLExpandsEnv(t *testing.T) {
	t.Setenv("WTTR_TEST_PORT", "8181")
	t.Setenv("WTTR_TEST_OPENMETEO_KEY", "test-openmeteo-key")

	path := filepath.Join(t.TempDir(), "config.yaml")
	err := os.WriteFile(path, []byte(`
geo:
  nominatim:
    - name: openmeteo
      type: openmeteo
      url: https://geocoding-api.open-meteo.com/v1/search
      token: "${WTTR_TEST_OPENMETEO_KEY}"
server:
  portHttp: ${WTTR_TEST_PORT}
`), 0o600)
	require.NoError(t, err)

	cfg, err := LoadFromYAML(path)
	require.NoError(t, err)
	require.Equal(t, 8181, cfg.Server.PortHTTP)
	require.Len(t, cfg.Geo.NominatimServers, 1)
	require.Equal(t, "test-openmeteo-key", cfg.Geo.NominatimServers[0].Token)
}

func TestLoadFromYAMLExpandsEnvDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	err := os.WriteFile(path, []byte(`
geo:
  nominatim:
    - name: openmeteo
      type: openmeteo
      url: https://geocoding-api.open-meteo.com/v1/search
      token: "${WTTR_TEST_MISSING_OPENMETEO_KEY:-}"
server:
  portHttp: ${WTTR_TEST_MISSING_PORT:-8080}
`), 0o600)
	require.NoError(t, err)

	cfg, err := LoadFromYAML(path)
	require.NoError(t, err)
	require.Equal(t, 8080, cfg.Server.PortHTTP)
	require.Len(t, cfg.Geo.NominatimServers, 1)
	require.Empty(t, cfg.Geo.NominatimServers[0].Token)
}

func TestLoadFromYAMLKeepsEnvValuesInsideOriginalScalar(t *testing.T) {
	injected := "secret\nserver:\n  portHttp: 1\nlogging:\n  interval: 1\n"
	t.Setenv("WTTR_TEST_INJECTED_OPENMETEO_KEY", injected)

	path := filepath.Join(t.TempDir(), "config.yaml")
	err := os.WriteFile(path, []byte(`
geo:
  nominatim:
    - name: openmeteo
      type: openmeteo
      url: https://geocoding-api.open-meteo.com/v1/search
      token: "${WTTR_TEST_INJECTED_OPENMETEO_KEY}"
server:
  portHttp: ${WTTR_TEST_MISSING_PORT:-8080}
`), 0o600)
	require.NoError(t, err)

	cfg, err := LoadFromYAML(path)
	require.NoError(t, err)
	require.Equal(t, 8080, cfg.Server.PortHTTP)
	require.Len(t, cfg.Geo.NominatimServers, 1)
	require.Equal(t, injected, cfg.Geo.NominatimServers[0].Token)
}
