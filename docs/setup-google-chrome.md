## Setup Instructions: Google Chrome
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
