package location

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/chubin/wttr.in/internal/types"
)

type locationOpenMeteo struct {
	Results []struct {
		Name      string  `json:"name"`
		Lat       float64 `json:"latitude"`
		Lon       float64 `json:"longitude"`
		Timezone  string  `json:"timezone"`
		Country   string  `json:"country"`
		Admin1    string  `json:"admin1"`
		Admin2    string  `json:"admin2"`
		Admin3    string  `json:"admin3"`
		Admin4    string  `json:"admin4"`
	} `json:"results"`
}

func (data *locationOpenMeteo) Query(n *Nominatim, location string) (*Location, error) {
	endpoint := fmt.Sprintf(
		"%s?name=%s&count=1&format=json",
		n.url, url.QueryEscape(location))

	if n.token != "" {
		endpoint += "&apikey=" + url.QueryEscape(n.token)
	}

	err := makeQuery(endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", n.name, err)
	}

	if len(data.Results) != 1 {
		return nil, fmt.Errorf("%w: %s: invalid response", types.ErrUpstream, n.name)
	}

	nl := data.Results[0]

	return &Location{
		Lat:      fmt.Sprint(nl.Lat),
		Lon:      fmt.Sprint(nl.Lon),
		Timezone: nl.Timezone,
		Fullname: openMeteoFullname(nl.Name, nl.Admin1, nl.Admin2, nl.Admin3, nl.Admin4, nl.Country),
	}, nil
}

func openMeteoFullname(parts ...string) string {
	var b strings.Builder
	var seen [6]string
	seenCount := 0

	for _, part := range parts {
		if part == "" {
			continue
		}

		duplicate := false
		for i := range seenCount {
			if part == seen[i] {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}

		if b.Len() != 0 {
			b.WriteString(", ")
		}
		b.WriteString(part)

		if seenCount < len(seen) {
			seen[seenCount] = part
			seenCount++
		}
	}

	return b.String()
}
