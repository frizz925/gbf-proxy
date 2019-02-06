## Setup Instructions: Proxy Auto-Configuration (PAC)
It's highly recommended to use Proxy Auto-Configuration (PAC) file if you use any other web proxy client tools. You can download the PAC file from this repository [using this link](https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/web/gbf-proxy.pac).

Or you can use the following PAC script instead (may be outdated).
```js
function FindProxyForURL(url, host) {
    if (dnsDomainIs(host, ".granbluefantasy.jp")) {
        return "PROXY gbf-proxy.kogane.moe:8088";
    }
    return "DIRECT";
}
```
