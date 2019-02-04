function FindProxyForURL(url, host) {
    if (dnsDomainIs(host, ".granbluefantasy.jp")) {
        return "PROXY gbf-proxy.kogane.moe:8088";
    }
    return "DIRECT";
}
