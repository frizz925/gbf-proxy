<br>
<p align="center"><img src="https://raw.githubusercontent.com/Frizz925/gbf-proxy/master/res/architecture-rev2.png"></p>
<br>

[![Build Status](https://travis-ci.org/Frizz925/gbf-proxy.svg?branch=master)](https://travis-ci.org/Frizz925/gbf-proxy)
[![codecov](https://codecov.io/gh/Frizz925/gbf-proxy/branch/master/graph/badge.svg)](https://codecov.io/gh/Frizz925/gbf-proxy)
[![License](https://img.shields.io/github/license/Frizz925/gbf-proxy.svg?style=flat)](https://github.com/Frizz925/gbf-proxy/blob/master/LICENSE)

# Granblue Proxy
Granblue caching proxy for in-game assets written in Golang and aimed to be blazingly fast!

## How to Use
The proxy is available on *gbf-proxy.kogane.moe* with the following ports:
- 80: HTTP proxy port (insecure, **not recommended**)
- 443: HTTPS proxy port (secure, recommended)
- 8088: Alternative HTTP proxy port

Although this is basically a public web proxy, it can only proxy your web traffic into **.granbluefantasy.jp** domains so make sure you set up your proxy rules properly, otherwise you'll get 403 Forbidden responses!

Setup instructions:
- [Google Chrome](https://github.com/Frizz925/gbf-proxy/blob/master/docs/setup-google-chrome.md)
- Android (TBA)
- iOS (TBA)
- [Proxy Auto-Configuration (PAC)](https://github.com/Frizz925/gbf-proxy/blob/master/docs/setup-pac.md)

*Readme still WIP*
