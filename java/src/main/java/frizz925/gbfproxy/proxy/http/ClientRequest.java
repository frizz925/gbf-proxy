package frizz925.gbfproxy.proxy.http;

import java.io.IOException;
import java.io.InputStream;
import java.io.PipedInputStream;
import java.io.PipedOutputStream;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.HashMap;
import java.util.Map;

public class ClientRequest {
    public static ClientRequest parse(String message) throws IOException, URISyntaxException {
        String header = message;
        String body = "";
        int bodyIdx = message.indexOf("\r\n\r\n");
        if (bodyIdx > 0) {
            header = message.substring(0, bodyIdx);
            body = message.substring(bodyIdx + 4);
        }

        ClientRequest req = new ClientRequest();
        req.message = message;
        req.pipe.write(body.getBytes());

        int headerIdx = header.indexOf("\r\n");
        if (headerIdx <= 0) {
            throw new IOException("Malformed HTTP header");
        }
        String requestLine = header.substring(0, headerIdx);
        String[] tokens = requestLine.split(" ");
        req.method = tokens[0].trim();
        req.uri = createUri(tokens[1].trim());
        req.version = tokens[2].split("/")[1].trim();

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
            req.requestHeaders.put(name, value);
        }

        return req;
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
    protected PipedOutputStream pipe;
    protected String message;

    private Map<String, String> requestHeaders;
    private InputStream inputStream;


    protected ClientRequest() throws IOException {
        this.requestHeaders = new HashMap<>();
        this.pipe = new PipedOutputStream();
        this.inputStream = new PipedInputStream(this.pipe);
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

    public InputStream getInputStream() {
        return inputStream;
    }

    public String getMessage() {
        return message;
    }
}