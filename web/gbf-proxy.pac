function FindProxyForURL(url, host) {
    var httpsProxy = "HTTPS gbf-proxy.kogane.moe:443";
    if (dnsDomainIs(host, ".granbluefantasy.jp")) {
        if (host.startsWith("game")) {
            return httpsProxy
        }
    } else if (dnsDomainIs(host, ".mbga.jp")) {
        if (host.startsWith("gbf.game")) {
            return httpsProxy
        }
    }
    return "DIRECT";
}
