package frizz925.gbfproxy.controller.server;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URI;
import java.net.URL;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.zip.GZIPInputStream;
import java.util.zip.GZIPOutputStream;

import javax.net.ssl.HttpsURLConnection;

import org.apache.commons.io.IOUtils;

import frizz925.gbfproxy.utils.Logger;
import io.undertow.server.HttpHandler;
import io.undertow.server.HttpServerExchange;
import io.undertow.util.HeaderValues;
import io.undertow.util.HttpString;

public class ControllerHandler implements HttpHandler {
    public static final String[] INTERCEPT_CONTENT_TYPES = new String[] {
        "text/",
        "application/",
    };

    @Override
    public void handleRequest(HttpServerExchange exchange) throws Exception {
        URI uri = URI.create(exchange.getRequestURL());
        String host = uri.getHost();
        if (host.endsWith(".mobage.jp")) {
            uri = URI.create(uri.toString().replace("http://", "https://"));
        }

        URL url = new URL(uri.toString());
        Logger.log("[Controller] Processing new request: " + uri.toString());
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        Logger.log("[Controller] Forwarding request...");
        forwardRequest(conn, exchange);
        Logger.log("[Controller] Returning response...");
        returnResponse(conn, exchange);
        Logger.log("[Controller] Request processed.");
    }

    protected void forwardRequest(HttpURLConnection conn, HttpServerExchange exchange) throws Exception {
        String method = exchange.getRequestMethod().toString();
        conn.setRequestMethod(method);
        for (HeaderValues header : exchange.getRequestHeaders()) {
            String key = header.getHeaderName().toString();
            for (String value : header) {
                conn.addRequestProperty(key, value);
            }
        }
        if (!method.equals("GET")) {
            conn.setDoOutput(true);
            IOUtils.copy(exchange.getInputStream(), conn.getOutputStream());
        }
    }

    protected void returnResponse(HttpURLConnection conn, HttpServerExchange exchange) throws IOException {
        int code = conn.getResponseCode();
        byte[] body = interceptResponse(conn, code >= 400 ? conn.getErrorStream() : conn.getInputStream());
        exchange.setStatusCode(code);
        exchange.getResponseHeaders().put(new HttpString("Content-Length"), body.length);
        Map<String, List<String>> headers = conn.getHeaderFields();
        for (Map.Entry<String, List<String>> entry : headers.entrySet()) {
            String key = entry.getKey();
            if (key == null || key.equals("Content-Length")) {
                continue;
            }
            HttpString name = new HttpString(key);
            List<String> values = entry.getValue();
            exchange.getResponseHeaders().addAll(name, values);
        }

        String method = conn.getRequestMethod();
        String host = conn.getURL().getHost();
        if (method.equals("OPTIONS")) {
            exchange.getResponseHeaders()
                .put(new HttpString("Access-Control-Allow-Origin"), "https://game.granbluefantasy.jp")
                .put(new HttpString("Access-Control-Allow-Methods"), "GET,PUT,POST,DELETE,PATCH,OPTIONS")
                .put(new HttpString("Access-Control-Allow-Headers"), "*")
                .put(new HttpString("Access-Control-Allow-Credentials"), "true")
                .put(new HttpString("Access-Control-Request-Headers"), "*");
        } else if (!host.equals("game.granbluefantasy.jp")) {
            exchange.getResponseHeaders()
                .put(new HttpString("Access-Control-Allow-Origin"), "https://game.granbluefantasy.jp")
                .put(new HttpString("Access-Control-Allow-Headers"), "*");
        }

        exchange.getOutputStream().write(body);
        exchange.getOutputStream().flush();
    }

    protected byte[] interceptResponse(HttpURLConnection conn, InputStream in) throws IOException {
        String contentLength = conn.getHeaderField("Content-Length");
        ByteArrayOutputStream baos = contentLength != null ? 
            new ByteArrayOutputStream(Integer.parseInt(contentLength)) :
            new ByteArrayOutputStream();
        String contentType = conn.getHeaderField("Content-Type");

        boolean valid = false;
        for (String prefix : INTERCEPT_CONTENT_TYPES) {
            if (contentType.startsWith(prefix)) {
                valid = true;
                break;
            }
        }
        if (!valid) {
            IOUtils.copy(in, baos);
            return baos.toByteArray();
        }

        String contentEncoding = conn.getHeaderField("Content-Encoding");
        boolean isGzip = contentEncoding != null && contentEncoding.equals("gzip");
        if (isGzip) {
            in = new GZIPInputStream(in);
        }

        IOUtils.copy(in, baos);
        byte[] bytes = baos.toString()
            .replaceAll("http://", "https://")
            .getBytes();

        baos.flush();
        baos.reset();
        if (isGzip) {
            OutputStream os = new GZIPOutputStream(baos);
            os.write(bytes);
            os.flush();
            os.close();
        } else {
            baos.write(bytes);
        }

        bytes = baos.toByteArray();
        return bytes;
    }
}