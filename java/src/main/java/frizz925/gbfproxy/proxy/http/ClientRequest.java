package frizz925.gbfproxy.proxy.http;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.HashMap;
import java.util.Map;

public class ClientRequest {
    public static final class Builder {
        private ClientRequest request;
        private ByteArrayOutputStream byteStream;

        protected Builder() {
            this.request = new ClientRequest();
            this.byteStream = new ByteArrayOutputStream();
        }

        public Builder setVersion(String version) {
            request.version = version;
            return this;
        }

        public Builder setUri(String uri) {
            return setUri(URI.create(uri));
        }

        public Builder setUri(URI uri) {
            request.uri = uri;
            return this;
        }

        public Builder setMethod(String method) {
            request.method = method;
            return this;
        }

        public Builder setRequestHeader(String name, String value) {
            request.getRequestHeaders().put(name, value);
            return this;
        }

        public Builder write(byte[] bytes) throws IOException {
            byteStream.write(bytes);
            return this;
        }

        public Builder write(byte[] bytes, int off, int len) throws IOException {
            byteStream.write(bytes, off, len);
            return this;
        }

        public ClientRequest build() {
            request.body = byteStream.toByteArray();
            return request;
        }
    }

    public static Builder builder() throws IOException {
        return new Builder();
    }

    public static ClientRequest parseSafe(byte[] payload) {
        try {
            return parse(payload);
        } catch (IOException e) {
            return null;
        } catch (URISyntaxException e) {
            return null;
        }
    }

    public static ClientRequest parse(byte[] payload) throws IOException, URISyntaxException {
        String message = new String(payload);
        if (!message.contains("\r\n\r\n")) {
            throw new IOException("Malformed HTTP header");
        }

        String header = message;
        String body = "";
        int bodyIdx = message.indexOf("\r\n\r\n");
        if (bodyIdx > 0) {
            header = message.substring(0, bodyIdx);
            body = message.substring(bodyIdx + 4);
        }

        int headerIdx = header.indexOf("\r\n");
        String requestLine = header.substring(0, headerIdx);
        String[] tokens = requestLine.split(" ");
        String method = tokens[0].trim();
        URI uri = createUri(tokens[1].trim());
        String version = tokens[2].split("/")[1].trim();

        Builder builder = builder()
            .setVersion(version)
            .setMethod(method)
            .setUri(uri)
            .write(body.getBytes());

        String[] headerLines = header.substring(headerIdx + 2)
            .trim()
            .split("\r\n");
        for (String line : headerLines) {
            int idx = line.indexOf(": ");
            if (idx <= 0) {
                continue;
            }
            String name = line.substring(0, idx);
            String value = line.substring(idx + 2);
            builder.setRequestHeader(name, value);
        }

        return builder.build();
    }

    protected static URI createUri(String raw) throws URISyntaxException {
        String scheme = "http";
        String path = "/";
        String fragment = null;
        String query = null;

        // Check for scheme
        int schemeIdx = raw.indexOf("://");
        if (schemeIdx > 0) {
            scheme = raw.substring(0, schemeIdx);
            schemeIdx += 3;
        } else {
            schemeIdx = 0;
        }

        // Check for fragment
        int fragmentIdx = raw.indexOf("#", schemeIdx);
        if (fragmentIdx > 0) {
            fragment = raw.substring(fragmentIdx);
        } else {
            fragmentIdx = raw.length();
        }

        // Check for query
        int queryIdx = raw.indexOf("?", schemeIdx);
        if (queryIdx > 0) {
            query = raw.substring(queryIdx + 1, fragmentIdx);
        } else {
            queryIdx = fragmentIdx;
        }

        // Check for path
        int pathIdx = raw.indexOf("/", schemeIdx);
        if (pathIdx > 0) {
            path = raw.substring(pathIdx, queryIdx);
        } else {
            pathIdx = queryIdx;
        }

        // Get the authority
        String authority = raw.substring(schemeIdx, pathIdx);
        return new URI(scheme, authority, path, query, fragment);
    }

    protected URI uri;
    protected String method;
    protected String version;
    protected byte[] body;

    private Map<String, String> requestHeaders;

    protected ClientRequest() {
        this.requestHeaders = new HashMap<>();
    }

    public URI getUri() {
        return uri;
    }

    public String getMethod() {
        return method;
    }

    public String getVersion() {
        return version;
    }

    public Map<String, String> getRequestHeaders() {
        return requestHeaders;
    }

    public byte[] getBody() {
        return body;
    }

    public String toString() {
        String result = String.format("%s %s HTTP/%s\r\n", method, uri.toString(), version);
        for (Map.Entry<String, String> entry : requestHeaders.entrySet()) {
            result += String.format("%s: %s\r\n", entry.getKey(), entry.getValue());
        }
        return result + "\r\n" + new String(body);
    }
}