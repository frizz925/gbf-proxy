package frizz925.gbfproxy.proxy.server;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.SocketAddress;
import java.net.URI;
import java.nio.ByteBuffer;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.SocketChannel;

import frizz925.gbfproxy.bootstrap.Application;
import frizz925.gbfproxy.proxy.channels.ReadHandler;
import frizz925.gbfproxy.proxy.http.ClientRequest;
import frizz925.gbfproxy.proxy.http.ServerResponse;
import frizz925.gbfproxy.utils.Logger;

public class TunnelHandler implements ReadHandler {
    public static final String[] ALLOWED_HOSTS = new String[] {
        ".granbluefantasy.jp",
        ".mobage.jp"
    };

    protected ByteBuffer buffer;
    protected SocketAddress address;
    protected SocketChannel peer;

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

        byte[] payload = readPayload(source);
        ClientRequest req = ClientRequest.parseSafe(payload);
        if (req != null) {
            tunnel(key, req);
        } else {
            if (peer != null) {
                tunnel(key, peer, payload);
            } else {
                reject(key, 400, "Bad Request");
            }
        }
    }

    public void tunnel(SelectionKey key, ClientRequest req) throws IOException {
        URI uri = req.getUri();
        String host = uri.getHost();
        if (host == null) {
            host = getHostFromHeader(req);
        }
        if (host == null) {
            reject(key, 400, "Bad Request");
            return;
        }

        boolean valid = false;
        for (String suffix : ALLOWED_HOSTS) {
            if (host.endsWith(suffix)) {
                valid = true;
                break;
            }
        }
        if (!valid) {
            reject(key, 403, "Forbidden");
            return;
        }

        if (req.getMethod().equals("CONNECT")) {
            peer = createPeer();
            accept(key);
            return;
        }
        if (!uri.getScheme().equals("https")) {
            String url = uri.toString();
            redirect(key, url.replace("http://", "https://"));
            if (peer != null) {
                peer.close();
                peer = null;
            }
            return;
        }
        if (peer == null) {
            peer = createPeer();
        }
        tunnel(key, peer, req.toString().getBytes());
    }

    public void tunnel(SelectionKey key, SocketChannel peer, byte[] payload) throws IOException {
        ByteBuffer bb = ByteBuffer.wrap(payload);
        while (bb.hasRemaining()) {
            peer.write(bb);
        }
        PeerHandler handler;
        Selector selector = key.selector();
        SocketChannel source = (SocketChannel) key.channel();

        handler = new PeerHandler(source);
        handler.read(peer.register(selector, SelectionKey.OP_READ, handler));
        handler = new PeerHandler(peer);
        handler.read(source.register(selector, SelectionKey.OP_READ, handler));
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

    public void redirect(SelectionKey key, String url) throws IOException {
        respond(key, ServerResponse.builder()
            .setCode(302)
            .setMessage("Found")
            .setResponseHeader("Location", url)
            .setResponseHeader("Connection", "Keep-Alive")
            .setResponseHeader("Content-Length", 0));
        Logger.log("[Proxy] Redirected connection");
    }

    public void respond(SelectionKey key, int code, String message) throws IOException {
        respond(key, ServerResponse.builder()
            .setCode(code)
            .setMessage(message));
    }

    public void respond(SelectionKey key, ServerResponse.Builder builder) throws IOException {
        builder.setResponseHeader("Server", Application.getFullName());
        respond(key, builder.build());
    }

    public void respond(SelectionKey key, ServerResponse response) throws IOException {
        respond(key, response.toString());
    }

    public void respond(SelectionKey key, String response) throws IOException {
        SocketChannel source = (SocketChannel) key.channel();
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

    private byte[] readPayload(SocketChannel source) throws IOException {
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
        return baos.toByteArray();
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

    private SocketChannel createPeer() throws IOException {
        SocketChannel peer = SocketChannel.open(this.address);
        peer.configureBlocking(false);
        return peer;
    }
}