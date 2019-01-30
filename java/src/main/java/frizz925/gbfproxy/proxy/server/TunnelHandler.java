package frizz925.gbfproxy.proxy.server;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.SocketAddress;
import java.nio.ByteBuffer;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.SocketChannel;

import frizz925.gbfproxy.bootstrap.Application;
import frizz925.gbfproxy.proxy.channels.ReadHandler;
import frizz925.gbfproxy.proxy.http.ClientRequest;
import frizz925.gbfproxy.utils.Logger;

public class TunnelHandler implements ReadHandler {
    protected ByteBuffer buffer;
    protected SocketAddress address;

    public TunnelHandler(SocketAddress address) {
        this.buffer = ByteBuffer.allocate(128);
        this.address = address;
    }

    @Override
    public void read(SelectionKey key) throws IOException {
        SocketChannel source = (SocketChannel) key.channel();
        if (!source.isOpen() || !source.isConnected()) {
            key.cancel();
            return;
        }

        try {
            String message = readToString(source);
            ClientRequest req = ClientRequest.parse(message);
            tunnel(key, req);
        } catch (Exception e) {
            Logger.error(e);
            reject(key, 400, "Bad Request");
        }
    }

    public void tunnel(SelectionKey key, ClientRequest req) throws IOException {
        String host = req.getUri().getHost();
        if (host == null) {
            host = getHostFromHeader(req);
        }
        if (host == null) {
            reject(key, 400, "Bad Request");
            return;
        }
        if (!host.endsWith(".granbluefantasy.jp")) {
            reject(key, 403, "Forbidden");
            return;
        }
        if (req.getMethod().equals("CONNECT")) {
            accept(key);
        }

        PeerHandler handler;
        Selector selector = key.selector();
        SocketChannel source = (SocketChannel) key.channel();
        SocketChannel peer = SocketChannel.open(this.address);
        peer.configureBlocking(false);

        String message = req.getMessage();
        ByteBuffer bb = ByteBuffer.wrap(message.getBytes());
        while (bb.hasRemaining()) {
            peer.write(bb);
        }

        handler = new PeerHandler(source);
        handler.read(peer.register(selector, SelectionKey.OP_READ, handler));
        handler = new PeerHandler(peer);
        handler.read(source.register(selector, SelectionKey.OP_READ, handler));
        key.cancel();
        Logger.log("[Proxy] Tunneled connection");
    }

    public void reject(SelectionKey key, int code, String message) throws IOException {
        respond(key, code, message);
        close(key);
        Logger.log("[Proxy] Rejected connection");
    }

    public void accept(SelectionKey key) throws IOException {
        respond(key, 200, "Connection Established");
    }

    public void respond(SelectionKey key, int code, String message) throws IOException {
        SocketChannel source = (SocketChannel) key.channel();
        String serverName = Application.getFullName();
        String response = String.join("\r\n", new String[] {
            "HTTP/1.1 " + code + " " + message,
            "Server: " + serverName,
            "\r\n"
        });
        ByteBuffer buffer = ByteBuffer.wrap(response.getBytes());
        while (buffer.hasRemaining()) {
            source.write(buffer);
        }
    }

    public void close(SelectionKey key) throws IOException {
        SocketChannel source = (SocketChannel) key.channel();
        source.close();
        key.cancel();
    }

    private String readToString(SocketChannel source) throws IOException {
        ByteArrayOutputStream baos = new ByteArrayOutputStream();
        while (source.isOpen() && source.isConnected()) {
            buffer.clear();
            int read = source.read(buffer);
            if (read <= 0) {
                break;
            }
            buffer.flip();
            baos.write(buffer.array(), buffer.position(), buffer.remaining());
        }
        return baos.toString();
    }

    private String getHostFromHeader(ClientRequest req) {
        String host = req.getRequestHeaders().get("Host");
        if (host == null) {
            return null;
        }
        int portIdx = host.indexOf(":");
        if (portIdx > 0) {
            return host.substring(0, portIdx);
        }
        return host;
    }
}