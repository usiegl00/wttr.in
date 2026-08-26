package location

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/chubin/wttr.in/internal/types"
)

type Nominatim struct {
	name  string
	url   string
	token string
	typ   string
}

type locationQuerier interface {
	Query(*Nominatim, string) (*Location, error)
}

func NewNominatim(name, typ, url, token string) *Nominatim {
	return &Nominatim{
		name:  name,
		url:   url,
		token: token,
		typ:   typ,
	}
}

func (n *Nominatim) Name() string {
	return fmt.Sprintf("%s (%s)", n.name, n.typ)
}

func (n *Nominatim) Query(location string) (*Location, error) {
	var data locationQuerier

	switch n.typ {
	case "iq":
		data = &locationIQ{}
	case "opencage":
		data = &locationOpenCage{}
	case "openmeteo":
		data = &locationOpenMeteo{}
	default:
		return nil, fmt.Errorf("%s: %w", n.name, types.ErrUnknownLocationService)
	}

	return data.Query(n, location)
}

func makeQuery(url string, result interface{}) error {
	var errResponse struct {
		Error  interface{} `json:"error"`
		Reason string      `json:"reason"`
	}

	log.Debugln("nominatim:", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &errResponse)
	if err == nil {
		switch upstreamErr := errResponse.Error.(type) {
		case string:
			if upstreamErr != "" {
				return fmt.Errorf("%w: %s", types.ErrUpstream, upstreamErr)
			}
		case bool:
			if upstreamErr {
				reason := errResponse.Reason
				if reason == "" {
					reason = "upstream error"
				}
				return fmt.Errorf("%w: %s", types.ErrUpstream, reason)
			}
		}
	}

	log.Debugln("nominatim: response: ", string(body))
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	return nil
}
