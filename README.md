<br>
<p align="center"><img src="https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/res/architecture-rev2.png"></p>
<br>

[![Build Status](https://travis-ci.org/Frizz925/gbf-proxy.svg?branch=master)](https://travis-ci.org/Frizz925/gbf-proxy)
[![codecov](https://codecov.io/gh/Frizz925/gbf-proxy/branch/master/graph/badge.svg)](https://codecov.io/gh/Frizz925/gbf-proxy)
[![License](https://img.shields.io/github/license/Frizz925/gbf-proxy.svg?style=flat)](https://github.com/Frizz925/gbf-proxy/blob/master/LICENSE)

# Granblue Proxy
Granblue caching proxy for in-game assets written in Golang and aimed to be blazingly fast!

## How to Use
The proxy is available on *gbf-proxy.kogane.moe* at ports 80, 443, and 8088. The server is hosted on Amazon Web Services EC2 in Singapore region. If you're located in SEA, this may still be useful for you. Otherwise, you're out of luck.

Although this is basically a public web proxy, it can only proxy your web traffic into **.granbluefantasy.jp** domains so make sure you set up your proxy rules properly, otherwise you'll get 403 Forbidden responses!

### Google Chrome
I highly recommend the Chrome Extension [Proxy SwitchyOmega](https://chrome.google.com/webstore/detail/proxy-switchyomega/padekgcemlokbadohgkifijomclgjgif?hl=en) to easily setup the proxy. It's free and it just works.

There are two ways to set this up:
- Using switch profile
- Using PAC profile

In this case, it's highly recommended to use the *switch profile* because it's just more configureable.

**1a. Creating switch profile**

First create a new *proxy profile* with the following configuration (you can use any other available ports).

![Proxy profile](https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/res/proxyswitch-1.png)

Finally, create a new *switch profile* with the following configuration.

![Switch profile](https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/res/proxyswitch-2.png)

**1b. Creating PAC profile**

Just create a new *PAC profile* with the following PAC URL: [https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/web/gbf-proxy.pac](https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/web/gbf-proxy.pac)

![Profile activation](https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/res/proxyswitch-3.png)

**2. Activating profile**

Make sure you have applied the changes to your profile. After that, simply just click the extension icon and select the profile that you just made.

![Profile activation](https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/res/proxyswitch-4.png)

### Proxy Auto-Configuration (PAC) file
It's highly recommended to use Proxy Auto-Configuration (PAC) file if you use any other web proxy client tools. You can download from this repository [using this link](https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/web/gbf-proxy.pac).

Or you can use the following PAC script instead (may be outdated).
```js
function FindProxyForURL(url, host) {
    if (dnsDomainIs(host, ".granbluefantasy.jp")) {
        return "PROXY gbf-proxy.kogane.moe:8088";
    }
    return "DIRECT";
}
```

*Readme still WIP*
