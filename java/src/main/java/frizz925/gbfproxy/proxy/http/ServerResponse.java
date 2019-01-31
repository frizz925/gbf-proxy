package frizz925.gbfproxy.proxy.http;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

public class ServerResponse {
    public static final class Builder {
        private ServerResponse response;
        private ByteArrayOutputStream byteStream;

        protected Builder() {
            this.response = new ServerResponse();
            this.byteStream = new ByteArrayOutputStream();
        }

        public Builder setVersion(String version) {
            response.version = version;
            return this;
        }

        public Builder setCode(int code) {
            response.code = code;
            return this;
        }

        public Builder setMessage(String message) {
            response.message = message;
            return this;
        }

        public Builder setResponseHeader(String name, String value) {
            response.responseHeaders.put(name, value);
            return this;
        }

        public Builder setResponseHeader(String name, int value) {
            return setResponseHeader(name, String.valueOf(value));
        }

        public Builder write(byte[] bytes) throws IOException {
            byteStream.write(bytes);
            return this;
        }

        public Builder write(byte[] bytes, int off, int len) throws IOException {
            byteStream.write(bytes, off, len);
            return this;
        }

        public ServerResponse build() {
            response.body = byteStream.toByteArray();
            return response;
        }
    }

    public static Builder builder() {
        return new Builder();
    }

    protected int code;
    protected String message;
    protected String version;
    protected byte[] body;

    private Map<String, String> responseHeaders;

    protected ServerResponse() {
        this.code = 200;
        this.message = "OK";
        this.version = "1.1";
        this.responseHeaders = new HashMap<>();

    }

    public int getCode() {
        return code;
    }

    public String getMessage() {
        return message;
    }

    public String getVersion() {
        return version;
    }

    public byte[] getBody() {
        return body;
    }

    public Map<String, String> getResponseHeaders() {
        return responseHeaders;
    }

    public String toString() {
        String result = String.format("HTTP/%s %d %s\r\n", version, code, message);
        for (Map.Entry<String, String> entry : responseHeaders.entrySet()) {
            result += String.format("%s: %s\r\n", entry.getKey(), entry.getValue());
        }
        return result + "\r\n" + new String(body);
    }
}