package spacedevs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// --- Output types ---

// Launch is a rocket launch event.
type Launch struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	NET         string `json:"net"`
	WindowStart string `json:"window_start"`
	WindowEnd   string `json:"window_end"`
	Probability int    `json:"probability"`
	Rocket      string `json:"rocket"`
	Mission     string `json:"mission"`
	MissionType string `json:"mission_type"`
	Orbit       string `json:"orbit"`
	Pad         string `json:"pad"`
	Location    string `json:"location"`
	Provider    string `json:"provider"`
}

// Astronaut is a space traveler record.
type Astronaut struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Agency      string `json:"agency"`
	Nationality string `json:"nationality"`
	DOB         string `json:"dob"`
	FlightCount int    `json:"flight_count"`
	Bio         string `json:"bio"`
}

// Agency is a space agency or organization.
type Agency struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Abbrev      string `json:"abbrev"`
	Type        string `json:"type"`
	CountryCode string `json:"country_code"`
	Website     string `json:"website"`
	Description string `json:"description"`
}

// Spacecraft is a spacecraft record.
type Spacecraft struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Agency      string `json:"agency"`
	Description string `json:"description"`
}

// --- Wire types ---

type wirePage[T any] struct {
	Count   int `json:"count"`
	Results []T `json:"results"`
}

type wireLaunch struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status struct {
		Name string `json:"name"`
	} `json:"status"`
	NET         string `json:"net"`
	WindowStart string `json:"window_start"`
	WindowEnd   string `json:"window_end"`
	Probability *int   `json:"probability"`
	Rocket      *struct {
		Configuration struct {
			Name string `json:"name"`
		} `json:"configuration"`
	} `json:"rocket"`
	Mission *struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Orbit *struct {
			Abbrev string `json:"abbrev"`
		} `json:"orbit"`
	} `json:"mission"`
	Pad *struct {
		Name     string `json:"name"`
		Location struct {
			Name string `json:"name"`
		} `json:"location"`
	} `json:"pad"`
	LaunchServiceProvider struct {
		Name string `json:"name"`
	} `json:"launch_service_provider"`
}

type wireAstronaut struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status struct {
		Name string `json:"name"`
	} `json:"status"`
	Agency struct {
		Name string `json:"name"`
	} `json:"agency"`
	Nationality string `json:"nationality"`
	DateOfBirth string `json:"date_of_birth"`
	FlightCount int    `json:"flights_count"`
	Bio         string `json:"bio"`
}

type wireAgency struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Abbrev      string `json:"abbrev"`
	Type        string `json:"type"`
	CountryCode string `json:"country_code"`
	Website     string `json:"website"`
	Description string `json:"description"`
}

type wireSpacecraft struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status struct {
		Name string `json:"name"`
	} `json:"status"`
	SpacecraftConfig *struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
		Agency struct {
			Name string `json:"name"`
		} `json:"agency"`
		Description string `json:"description"`
	} `json:"spacecraft_config"`
}

// --- Client methods ---

// ListUpcoming returns upcoming rocket launches.
func (c *Client) ListUpcoming(ctx context.Context, limit int, search string) ([]*Launch, error) {
	return c.listLaunches(ctx, "/launch/upcoming/", limit, search)
}

// ListLaunches returns all launches (past and upcoming).
func (c *Client) ListLaunches(ctx context.Context, limit int, search string) ([]*Launch, error) {
	return c.listLaunches(ctx, "/launch/", limit, search)
}

func (c *Client) listLaunches(ctx context.Context, path string, limit int, search string) ([]*Launch, error) {
	q := url.Values{"format": {"json"}}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if search != "" {
		q.Set("search", search)
	}
	endpoint := c.BaseURL + path + "?" + q.Encode()
	body, err := c.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	var page wirePage[wireLaunch]
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("parse launches: %w", err)
	}
	out := make([]*Launch, 0, len(page.Results))
	for _, w := range page.Results {
		out = append(out, convertLaunch(w))
	}
	return out, nil
}

func convertLaunch(w wireLaunch) *Launch {
	l := &Launch{
		ID:          w.ID,
		Name:        w.Name,
		Status:      w.Status.Name,
		NET:         w.NET,
		WindowStart: w.WindowStart,
		WindowEnd:   w.WindowEnd,
		Provider:    w.LaunchServiceProvider.Name,
	}
	if w.Probability != nil {
		l.Probability = *w.Probability
	}
	if w.Rocket != nil {
		l.Rocket = w.Rocket.Configuration.Name
	}
	if w.Mission != nil {
		l.Mission = w.Mission.Name
		l.MissionType = w.Mission.Type
		if w.Mission.Orbit != nil {
			l.Orbit = w.Mission.Orbit.Abbrev
		}
	}
	if w.Pad != nil {
		l.Pad = w.Pad.Name
		l.Location = w.Pad.Location.Name
	}
	return l
}

// ListAstronauts returns astronauts/space travelers.
func (c *Client) ListAstronauts(ctx context.Context, limit int, search string) ([]*Astronaut, error) {
	q := url.Values{"format": {"json"}}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if search != "" {
		q.Set("search", search)
	}
	endpoint := c.BaseURL + "/astronaut/?" + q.Encode()
	body, err := c.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	var page wirePage[wireAstronaut]
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("parse astronauts: %w", err)
	}
	out := make([]*Astronaut, 0, len(page.Results))
	for _, w := range page.Results {
		out = append(out, &Astronaut{
			ID:          w.ID,
			Name:        w.Name,
			Status:      w.Status.Name,
			Agency:      w.Agency.Name,
			Nationality: w.Nationality,
			DOB:         w.DateOfBirth,
			FlightCount: w.FlightCount,
			Bio:         w.Bio,
		})
	}
	return out, nil
}

// ListAgencies returns space agencies and organizations.
func (c *Client) ListAgencies(ctx context.Context, limit int, agencyType string) ([]*Agency, error) {
	q := url.Values{"format": {"json"}}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if agencyType != "" {
		q.Set("type", agencyType)
	}
	endpoint := c.BaseURL + "/agency/?" + q.Encode()
	body, err := c.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	var page wirePage[wireAgency]
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("parse agencies: %w", err)
	}
	out := make([]*Agency, 0, len(page.Results))
	for _, w := range page.Results {
		out = append(out, &Agency{
			ID:          w.ID,
			Name:        w.Name,
			Abbrev:      w.Abbrev,
			Type:        w.Type,
			CountryCode: w.CountryCode,
			Website:     w.Website,
			Description: w.Description,
		})
	}
	return out, nil
}

// ListSpacecraft returns spacecraft records.
func (c *Client) ListSpacecraft(ctx context.Context, limit int, search string) ([]*Spacecraft, error) {
	q := url.Values{"format": {"json"}}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if search != "" {
		q.Set("search", search)
	}
	endpoint := c.BaseURL + "/spacecraft/?" + q.Encode()
	body, err := c.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	var page wirePage[wireSpacecraft]
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("parse spacecraft: %w", err)
	}
	out := make([]*Spacecraft, 0, len(page.Results))
	for _, w := range page.Results {
		sc := &Spacecraft{
			ID:     w.ID,
			Name:   w.Name,
			Status: w.Status.Name,
		}
		if w.SpacecraftConfig != nil {
			sc.Type = w.SpacecraftConfig.Type.Name
			sc.Agency = w.SpacecraftConfig.Agency.Name
			sc.Description = w.SpacecraftConfig.Description
		}
		out = append(out, sc)
	}
	return out, nil
}
