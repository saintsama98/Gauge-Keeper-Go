package config

import (
	"strings"
	"testing"
	"time"

	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

const idHex = "0x" + "aa00000000000000000000000000000000000000000000000000000000000001"

func TestParseStrategyID(t *testing.T) {
	id, ok := parseStrategyID(idHex)
	if !ok || id[0] != 0xaa || id[31] != 0x01 {
		t.Fatalf("parseStrategyID(%q) = %x, %v", idHex, id, ok)
	}
	if _, ok := parseStrategyID(strings.TrimPrefix(idHex, "0x")); !ok {
		t.Fatal("bare hex must parse")
	}
	for _, bad := range []string{"", "0x12", idHex + "00", "0x" + strings.Repeat("zz", 32)} {
		if _, ok := parseStrategyID(bad); ok {
			t.Fatalf("parseStrategyID(%q) accepted malformed input", bad)
		}
	}
}

func TestParseListsSkipMalformedEntries(t *testing.T) {
	ids := parseStrategyIDs(idHex + ", nope ," + idHex)
	if len(ids) != 2 {
		t.Fatalf("parseStrategyIDs kept %d entries, want 2", len(ids))
	}
	if parseStrategyIDs("  ") != nil {
		t.Fatal("empty list must parse to nil")
	}

	m := parseStringMap(idHex + "=market-a, broken, nope=x")
	want, _ := parseStrategyID(idHex)
	if len(m) != 1 || m[want] != "market-a" {
		t.Fatalf("parseStringMap = %v", m)
	}
	if parseStringMap("broken") != nil {
		t.Fatal("a map with no valid entry must parse to nil")
	}

	pm := parsePendleMarkets(idHex + "=0xmarket:0xpt:900, " + idHex + "=0xmarket:0xpt")
	if got := pm[want]; got.Market != "0xmarket" || got.PT != "0xpt" || got.Duration != 900 {
		t.Fatalf("parsePendleMarkets = %+v", pm)
	}
	if parsePendleMarkets(idHex+"=a:b:notanumber") != nil {
		t.Fatal("a market with a bad duration must be skipped")
	}
}

func TestLoadDefaultsAndOverrides(t *testing.T) {
	t.Setenv("KEEPER_CHAIN_ID", "1")
	t.Setenv("KEEPER_DRY_RUN", "false")
	t.Setenv("KEEPER_POLL_INTERVAL", "7s")
	t.Setenv("KEEPER_GUARD_INTERVAL", "not-a-duration")
	t.Setenv("KEEPER_AAVE_STRATEGY_IDS", idHex)

	c := Load()
	if c.ChainID != 1 || c.DryRun || c.PollInterval != 7*time.Second {
		t.Fatalf("overrides not applied: chain %d dry %v poll %s", c.ChainID, c.DryRun, c.PollInterval)
	}
	if c.GuardInterval != 5*time.Minute {
		t.Fatalf("a malformed duration must fall back to the default, got %s", c.GuardInterval)
	}
	if c.Brain != "stub" || c.RiskProvider != "static" {
		t.Fatalf("defaults changed: brain %q risk %q", c.Brain, c.RiskProvider)
	}
	if len(c.AaveStrategyIDs) != 1 || c.AaveStrategyIDs[0] == (types.StrategyID{}) {
		t.Fatalf("strategy ids not parsed: %v", c.AaveStrategyIDs)
	}
}
