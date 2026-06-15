package spacedevs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(ts *httptest.Server) *Client {
	c := NewClient()
	c.BaseURL = ts.URL
	c.Rate = 0 // no pacing in tests
	return c
}

func mustEncode(t *testing.T, w http.ResponseWriter, v interface{}) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient()
	c.Rate = 0

	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("recovered"))
	}))
	defer srv.Close()

	c := NewClient()
	c.Rate = 0
	c.Retries = 5

	start := time.Now()
	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "recovered" {
		t.Errorf("body = %q after retries", body)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestListUpcoming(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mustEncode(t, w, map[string]interface{}{
			"count": 366,
			"results": []map[string]interface{}{
				{
					"id":   "abc-123",
					"name": "Falcon 9 | Starlink",
					"status": map[string]interface{}{
						"name": "Go for Launch",
					},
					"net":          "2026-07-01T10:00:00Z",
					"window_start": "2026-07-01T10:00:00Z",
					"window_end":   "2026-07-01T12:00:00Z",
					"probability":  85,
					"rocket": map[string]interface{}{
						"configuration": map[string]interface{}{
							"name": "Falcon 9",
						},
					},
					"mission": map[string]interface{}{
						"name": "Starlink Group 10-1",
						"type": "Communications",
						"orbit": map[string]interface{}{
							"abbrev": "LEO",
						},
					},
					"pad": map[string]interface{}{
						"name": "SLC-40",
						"location": map[string]interface{}{
							"name": "Cape Canaveral, FL",
						},
					},
					"launch_service_provider": map[string]interface{}{
						"name": "SpaceX",
					},
				},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	launches, err := c.ListUpcoming(context.Background(), 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(launches) != 1 {
		t.Fatalf("got %d launches, want 1", len(launches))
	}
	l := launches[0]
	if l.ID != "abc-123" {
		t.Errorf("ID = %q, want abc-123", l.ID)
	}
	if l.Name != "Falcon 9 | Starlink" {
		t.Errorf("Name = %q, want Falcon 9 | Starlink", l.Name)
	}
	if l.Status != "Go for Launch" {
		t.Errorf("Status = %q, want Go for Launch", l.Status)
	}
	if l.Provider != "SpaceX" {
		t.Errorf("Provider = %q, want SpaceX", l.Provider)
	}
	if l.Rocket != "Falcon 9" {
		t.Errorf("Rocket = %q, want Falcon 9", l.Rocket)
	}
	if l.Probability != 85 {
		t.Errorf("Probability = %d, want 85", l.Probability)
	}
	if l.Orbit != "LEO" {
		t.Errorf("Orbit = %q, want LEO", l.Orbit)
	}
}

func TestListLaunches(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mustEncode(t, w, map[string]interface{}{
			"count": 7900,
			"results": []map[string]interface{}{
				{
					"id":   "hist-001",
					"name": "Apollo 11",
					"status": map[string]interface{}{
						"name": "Launch Successful",
					},
					"net": "1969-07-16T13:32:00Z",
					"launch_service_provider": map[string]interface{}{
						"name": "NASA",
					},
				},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	launches, err := c.ListLaunches(context.Background(), 5, "Apollo")
	if err != nil {
		t.Fatal(err)
	}
	if len(launches) != 1 {
		t.Fatalf("got %d launches, want 1", len(launches))
	}
	if launches[0].Name != "Apollo 11" {
		t.Errorf("Name = %q, want Apollo 11", launches[0].Name)
	}
}

func TestListAstronauts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mustEncode(t, w, map[string]interface{}{
			"count": 858,
			"results": []map[string]interface{}{
				{
					"id":   42,
					"name": "Neil Armstrong",
					"status": map[string]interface{}{
						"name": "Deceased",
					},
					"agency": map[string]interface{}{
						"name": "NASA",
					},
					"nationality":   "American",
					"date_of_birth": "1930-08-05",
					"flights_count": 2,
					"bio":           "First human to walk on the Moon.",
				},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	astronauts, err := c.ListAstronauts(context.Background(), 5, "Armstrong")
	if err != nil {
		t.Fatal(err)
	}
	if len(astronauts) != 1 {
		t.Fatalf("got %d astronauts, want 1", len(astronauts))
	}
	a := astronauts[0]
	if a.ID != 42 {
		t.Errorf("ID = %d, want 42", a.ID)
	}
	if a.Name != "Neil Armstrong" {
		t.Errorf("Name = %q, want Neil Armstrong", a.Name)
	}
	if a.Agency != "NASA" {
		t.Errorf("Agency = %q, want NASA", a.Agency)
	}
	if a.FlightCount != 2 {
		t.Errorf("FlightCount = %d, want 2", a.FlightCount)
	}
}

func TestListAgencies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mustEncode(t, w, map[string]interface{}{
			"count": 50,
			"results": []map[string]interface{}{
				{
					"id":           1,
					"name":         "NASA",
					"abbrev":       "NASA",
					"type":         "Government",
					"country_code": "USA",
					"website":      "https://www.nasa.gov/",
					"description":  "The National Aeronautics and Space Administration.",
				},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	agencies, err := c.ListAgencies(context.Background(), 10, "Government")
	if err != nil {
		t.Fatal(err)
	}
	if len(agencies) != 1 {
		t.Fatalf("got %d agencies, want 1", len(agencies))
	}
	ag := agencies[0]
	if ag.Name != "NASA" {
		t.Errorf("Name = %q, want NASA", ag.Name)
	}
	if ag.Type != "Government" {
		t.Errorf("Type = %q, want Government", ag.Type)
	}
	if ag.CountryCode != "USA" {
		t.Errorf("CountryCode = %q, want USA", ag.CountryCode)
	}
}

func TestListSpacecraft(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mustEncode(t, w, map[string]interface{}{
			"count": 606,
			"results": []map[string]interface{}{
				{
					"id":   7,
					"name": "Dragon C201",
					"status": map[string]interface{}{
						"name": "Active",
					},
					"spacecraft_config": map[string]interface{}{
						"type": map[string]interface{}{
							"name": "Capsule",
						},
						"agency": map[string]interface{}{
							"name": "SpaceX",
						},
						"description": "SpaceX Dragon cargo capsule.",
					},
				},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	spacecraft, err := c.ListSpacecraft(context.Background(), 5, "Dragon")
	if err != nil {
		t.Fatal(err)
	}
	if len(spacecraft) != 1 {
		t.Fatalf("got %d spacecraft, want 1", len(spacecraft))
	}
	sc := spacecraft[0]
	if sc.Name != "Dragon C201" {
		t.Errorf("Name = %q, want Dragon C201", sc.Name)
	}
	if sc.Agency != "SpaceX" {
		t.Errorf("Agency = %q, want SpaceX", sc.Agency)
	}
	if sc.Type != "Capsule" {
		t.Errorf("Type = %q, want Capsule", sc.Type)
	}
}
