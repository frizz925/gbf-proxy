package frizz925.gbfproxy.proxy.server;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.SocketAddress;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;


import frizz925.gbfproxy.proxy.channels.AcceptHandler;
import frizz925.gbfproxy.utils.Logger;

public class ProxyHandler implements AcceptHandler {
    protected SocketAddress address;

    public ProxyHandler(int backendPort) {
        this.address = new InetSocketAddress("localhost", backendPort);
    }

    @Override
    public void accept(SelectionKey key) throws IOException {
        Logger.log("[Proxy] New connection");
        Selector selector = key.selector();
        ServerSocketChannel server = (ServerSocketChannel) key.channel();
        SocketChannel client = server.accept();
        client.configureBlocking(false);
        TunnelHandler handler = new TunnelHandler(this.address);
        client.register(selector, SelectionKey.OP_READ, handler);
    }
}