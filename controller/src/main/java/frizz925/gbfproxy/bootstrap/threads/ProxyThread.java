package frizz925.gbfproxy.bootstrap.threads;

import frizz925.gbfproxy.common.server.ServerInterface;
import frizz925.gbfproxy.proxy.server.ProxyServer;

public class ProxyThread extends ServerThread {
    public ProxyThread(int port) {
        super(port);
    }

    public ProxyThread(String host, int port) {
        super(host, port);
    }

    @Override
    public String getServerName() {
        return "Proxy";
    }

    @Override
    public ServerInterface getServer() {
        return new ProxyServer();
    }
}