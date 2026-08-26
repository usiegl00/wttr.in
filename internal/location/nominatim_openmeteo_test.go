package location_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chubin/wttr.in/internal/location"
	"github.com/chubin/wttr.in/internal/types"
)

func TestOpenMeteoQuery(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" {
			t.Errorf("path = %q, want /v1/search", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("name") != "New York, NY" {
			t.Errorf("name query = %q, want New York, NY", query.Get("name"))
		}
		if query.Get("count") != "1" {
			t.Errorf("count query = %q, want 1", query.Get("count"))
		}
		if query.Get("format") != "json" {
			t.Errorf("format query = %q, want json", query.Get("format"))
		}
		if query.Get("apikey") != "test-token" {
			t.Errorf("apikey query = %q, want test-token", query.Get("apikey"))
		}

		if _, err := w.Write([]byte(`{
			"results": [{
				"name": "Zurich",
				"latitude": 47.36667,
				"longitude": 8.55,
				"timezone": "Europe/Zurich",
				"country": "Switzerland",
				"admin1": "Zurich"
			}]
		}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	provider := location.NewNominatim("openmeteo", "openmeteo", server.URL+"/v1/search", "test-token")

	loc, err := provider.Query("New York, NY")
	require.NoError(t, err)
	require.Equal(t, &location.Location{
		Lat:      "47.36667",
		Lon:      "8.55",
		Timezone: "Europe/Zurich",
		Fullname: "Zurich, Switzerland",
	}, loc)
}

func TestOpenMeteoQueryRejectsEmptyResults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"results": []}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	provider := location.NewNominatim("openmeteo", "openmeteo", server.URL, "")

	loc, err := provider.Query("missing")
	require.Nil(t, loc)
	require.ErrorIs(t, err, types.ErrUpstream)
}

func TestOpenMeteoQueryReturnsReasonError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"error": true, "reason": "Parameter count must be between 1 and 100."}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	provider := location.NewNominatim("openmeteo", "openmeteo", server.URL, "")

	loc, err := provider.Query("bad")
	require.Nil(t, loc)
	require.ErrorIs(t, err, types.ErrUpstream)
	require.ErrorContains(t, err, "Parameter count must be between 1 and 100.")
}
