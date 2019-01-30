package frizz925.gbfproxy.bootstrap.threads;

import frizz925.gbfproxy.common.server.ServerInterface;
import frizz925.gbfproxy.controller.server.ControllerServer;

public class ControllerThread extends ServerThread {
    public ControllerThread(int port) {
        super(port);
    }

    public ControllerThread(String host, int port) {
        super(host, port);
    }

    @Override
    public String getServerName() {
        return "Controller";
    }

    @Override
    public ServerInterface getServer() {
        return new ControllerServer();
    }
}