package frizz925.gbfproxy.common.server;

public interface ServerInterface {
    public void start(int port);
    public void start(String host, int port);
}