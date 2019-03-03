package acl

import "strings"

func IsGameDomain(hostname string) bool {
	if IsGameAssetsDomain(hostname) {
		return true
	} else if hostname == "game.granbluefantasy.jp" {
		return true
	} else if hostname == "gbf.game.mbga.jp" {
		return true
	}
	return false
}

func IsGameAssetsDomain(hostname string) bool {
	if strings.HasPrefix(hostname, "game-a") && strings.HasSuffix(hostname, ".granbluefantasy.jp") {
		return true
	} else if strings.HasPrefix(hostname, "gbf.game-a") && strings.HasSuffix(hostname, ".mbga.jp") {
		return true
	}
	return false
}
