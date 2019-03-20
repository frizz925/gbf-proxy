package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		assert.Truef(t, IsGameAssetsDomain(h), "%s should be game assets domain", h)
		assert.Truef(t, IsGameDomain(h), "%s should be game domain", h)
	}

	for _, h := range allowed {
		assert.Falsef(t, IsGameAssetsDomain(h), "%s should NOT be game assets domain", h)
		assert.Truef(t, IsGameDomain(h), "%s should be game domain", h)
	}

	for _, h := range forbidden {
		assert.Falsef(t, IsGameDomain(h), "%s should NOT be game domain", h)
	}
}
