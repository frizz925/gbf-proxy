package acl

import "testing"

func TestDomains(t *testing.T) {
	assetHostnames := []string{
		"game-a.granbluefantasy.jp",
		"game-a1.granbluefantasy.jp",
		"gbf.game-a.mbga.jp",
		"gbf.game-a1.mbga.jp",
	}
	allowed := []string{
		"game.granbluefantasy.jp",
		"gbf.game.mbga.jp",
	}
	forbidden := []string{
		"google.com",
		"connect.mobage.jp",
		"game.notgranbluefantasy.jp",
		"notgame.granbluefantasy.jp",
		"www.granbluefantasy.jp",
	}

	for _, h := range assetHostnames {
		if !IsGameAssetsDomain(h) {
			t.Fatalf("%s should be game assets domain", h)
		}
		if !IsGameDomain(h) {
			t.Fatalf("%s should be game domain", h)
		}
	}

	for _, h := range allowed {
		if IsGameAssetsDomain(h) {
			t.Fatalf("%s should NOT be game assets domain", h)
		}
		if !IsGameDomain(h) {
			t.Fatalf("%s should be game domain", h)
		}
	}

	for _, h := range forbidden {
		if IsGameDomain(h) {
			t.Fatalf("%s should NOT be game domain", h)
		}
	}
}
