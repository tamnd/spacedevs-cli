package spacedevs

import (
	"context"

	"github.com/tamnd/any-cli/kit"
)

// domain.go exposes The Space Devs as a kit Domain: a driver that a multi-domain
// host (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/spacedevs-cli/spacedevs"
//
// exactly as a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// spacedevs:// URIs by routing to the operations Register installs. The same
// Domain also builds the standalone spacedevs binary (see cli.NewApp), so the
// binary and a host share one source of truth.
func init() { kit.Register(Domain{}) }

// Domain is the Space Devs driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against, and
// the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "spacedevs",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "spacedevs",
			Short:  "Space launch and astronaut data from The Space Devs Launch Library 2",
			Long: `Space launch and astronaut data from The Space Devs Launch Library 2.

spacedevs reads public data from ll.thespacedevs.com (no API key required),
shapes it into clean records, and prints output that pipes into the rest of
your tools. Browse upcoming launches, search the astronaut database, explore
spacecraft and agencies.`,
			Site: Host,
			Repo: "https://github.com/tamnd/spacedevs-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "upcoming", Group: "read", List: true,
		Summary: "Upcoming rocket launches"}, listUpcoming)
	kit.Handle(app, kit.OpMeta{Name: "launches", Group: "read", List: true,
		Summary: "All launches (past and upcoming)"}, listLaunches)
	kit.Handle(app, kit.OpMeta{Name: "astronauts", Group: "read", List: true,
		Summary: "Space travelers database"}, listAstronauts)
	kit.Handle(app, kit.OpMeta{Name: "agencies", Group: "read", List: true,
		Summary: "Space agencies and organizations"}, listAgencies)
	kit.Handle(app, kit.OpMeta{Name: "spacecraft", Group: "read", List: true,
		Summary: "Spacecraft catalog"}, listSpacecraft)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := NewClient()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.HTTP.Timeout = cfg.Timeout
	}
	return c, nil
}

// --- inputs ---

type listInput struct {
	Search string  `kit:"flag" help:"search term"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type agencyInput struct {
	Search string  `kit:"flag" help:"search term"`
	Type   string  `kit:"flag" help:"agency type (Government, Commercial, Educational)"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listUpcoming(ctx context.Context, in listInput, emit func(*Launch) error) error {
	launches, err := in.Client.ListUpcoming(ctx, in.Limit, in.Search)
	if err != nil {
		return err
	}
	for _, l := range launches {
		if err := emit(l); err != nil {
			return err
		}
	}
	return nil
}

func listLaunches(ctx context.Context, in listInput, emit func(*Launch) error) error {
	launches, err := in.Client.ListLaunches(ctx, in.Limit, in.Search)
	if err != nil {
		return err
	}
	for _, l := range launches {
		if err := emit(l); err != nil {
			return err
		}
	}
	return nil
}

func listAstronauts(ctx context.Context, in listInput, emit func(*Astronaut) error) error {
	astronauts, err := in.Client.ListAstronauts(ctx, in.Limit, in.Search)
	if err != nil {
		return err
	}
	for _, a := range astronauts {
		if err := emit(a); err != nil {
			return err
		}
	}
	return nil
}

func listAgencies(ctx context.Context, in agencyInput, emit func(*Agency) error) error {
	agencies, err := in.Client.ListAgencies(ctx, in.Limit, in.Type)
	if err != nil {
		return err
	}
	for _, a := range agencies {
		if err := emit(a); err != nil {
			return err
		}
	}
	return nil
}

func listSpacecraft(ctx context.Context, in listInput, emit func(*Spacecraft) error) error {
	spacecraft, err := in.Client.ListSpacecraft(ctx, in.Limit, in.Search)
	if err != nil {
		return err
	}
	for _, s := range spacecraft {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}
