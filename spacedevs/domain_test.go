package spacedevs

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the domain info and host wiring,
// which need no network.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "spacedevs" {
		t.Errorf("Scheme = %q, want spacedevs", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "spacedevs" {
		t.Errorf("Identity.Binary = %q, want spacedevs", info.Identity.Binary)
	}
}

func TestDomainRegister(t *testing.T) {
	app := kit.New(Domain{}.Info().Identity)
	Domain{}.Register(app)
	// Register must not panic; reaching here means it succeeded.
}
