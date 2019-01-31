package frizz925.gbfproxy.controller.server;

import frizz925.gbfproxy.common.server.ServerInterface;
import io.undertow.Undertow;
import io.undertow.UndertowOptions;
import io.undertow.server.handlers.BlockingHandler;

public class ControllerServer implements ServerInterface {
    public void start(int port) {
        start("0.0.0.0", port);
    }

    public void start(String host, int port) {
        Undertow server = Undertow.builder()
            .addHttpListener(port, host)
            .setServerOption(UndertowOptions.ENABLE_HTTP2, true)
            .setHandler(new BlockingHandler(new ControllerHandler()))
            .build();
        server.start();
    }
}