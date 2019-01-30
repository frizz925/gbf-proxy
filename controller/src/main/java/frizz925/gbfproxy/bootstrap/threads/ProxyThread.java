package frizz925.gbfproxy.bootstrap.threads;

import frizz925.gbfproxy.common.server.ServerInterface;
import frizz925.gbfproxy.proxy.server.ProxyServer;

public class ProxyThread extends ServerThread {
    public int backendPort;

    public ProxyThread(int port, int backendPort) {
        this("localhost", port, backendPort);
    }

    public ProxyThread(String host, int port, int backendPort) {
        super(host, port);
        this.backendPort = backendPort;
    }

    @Override
    public String getServerName() {
        return "Proxy";
    }

    @Override
    public ServerInterface getServer() {
        return new ProxyServer(this.backendPort);
    }
}