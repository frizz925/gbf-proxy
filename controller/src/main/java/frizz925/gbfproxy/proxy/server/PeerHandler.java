package frizz925.gbfproxy.proxy.server;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.channels.SelectionKey;
import java.nio.channels.SocketChannel;

import frizz925.gbfproxy.proxy.channels.ReadHandler;

public class PeerHandler implements ReadHandler {
    protected SocketChannel peer;
    protected ByteBuffer buffer;

    public PeerHandler(SocketChannel peer) {
        this.peer = peer;
        this.buffer = ByteBuffer.allocate(2048);
    }

    @Override
    public void read(SelectionKey key) throws IOException {
        SocketChannel source = (SocketChannel) key.channel();
        while (true) {
            buffer.clear();
            int read = source.read(buffer);
            if (read <= 0) {
                break;
            }
            buffer.flip();
            while (buffer.hasRemaining()) {
                peer.write(buffer);
            }
        }
    }
}