package frizz925.gbfproxy.proxy.server;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.util.Iterator;

import frizz925.gbfproxy.common.server.ServerInterface;
import frizz925.gbfproxy.proxy.channels.AcceptHandler;
import frizz925.gbfproxy.proxy.channels.ReadHandler;
import frizz925.gbfproxy.utils.Logger;

public class ProxyServer implements ServerInterface {
    protected int backendPort;
    protected boolean running = false;

    public ProxyServer(int backendPort) {
        this.backendPort = backendPort;
    }

    @Override
    public void start(int port) {
        start("localhost", port);
    }

    @Override
    public void start(String host, int port) {
        try {
            startServer(host, port);
        } catch (Exception e) {
            Logger.error(e);
        }
    }

    public void startServer(String host, int port) throws IOException {
        Selector selector = Selector.open();
        ServerSocketChannel server = ServerSocketChannel.open();
        ProxyHandler handler = new ProxyHandler(this.backendPort);
        server.bind(new InetSocketAddress(host, port));
        server.configureBlocking(false);
        server.register(selector, SelectionKey.OP_ACCEPT, handler);
        this.running = true;
        while (this.running) {
            try {
                select(selector);
            } catch (Exception e) {
                Logger.error(e);
            }
        } 
    }

    public void select(Selector selector) throws IOException {
        if (selector.select() <= 0) {
            return;
        }
        Iterator<SelectionKey> it = selector.selectedKeys().iterator();
        while (it.hasNext()) {
            SelectionKey key = it.next();
            Object handler = key.attachment();
            if (key.isAcceptable() && (handler instanceof AcceptHandler)) {
                ((AcceptHandler) handler).accept(key);
            }
            if (key.isReadable() && (handler instanceof ReadHandler)) {
                ((ReadHandler) handler).read(key);
            }
            it.remove();
        }
    }
}