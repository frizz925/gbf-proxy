package frizz925.gbfproxy.bootstrap.threads;

import frizz925.gbfproxy.common.server.ServerInterface;

abstract public class ServerThread extends Thread {
    public String host;
    public int port;

    public ServerThread(int port) {
        this("localhost", port);
    }

    public ServerThread(String host, int port) {
        super();
        this.host = host;
        this.port = port;
        this.setName(this.getServerName() + " Thread");
    }

    @Override
    public void run() {
        this.getServer().start(this.host, this.port);
    }

    abstract public String getServerName();
    abstract public ServerInterface getServer();
}