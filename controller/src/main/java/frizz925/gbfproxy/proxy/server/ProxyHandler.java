package frizz925.gbfproxy.proxy.server;

import java.io.IOException;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;

import frizz925.gbfproxy.proxy.channels.AcceptHandler;

public class ProxyHandler implements AcceptHandler {
    @Override
    public void accept(SelectionKey key) throws IOException {
        Selector selector = key.selector();
        ServerSocketChannel server = (ServerSocketChannel) key.channel();
        SocketChannel client = server.accept();
        client.register(selector, SelectionKey.OP_READ);
    }
}