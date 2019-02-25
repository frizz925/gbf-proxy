function FindProxyForURL(url, host) {
    if (dnsDomainIs(host, ".granbluefantasy.jp")) {
        return "HTTPS gbf-proxy.kogane.moe:443";
    }
    return "DIRECT";
}
