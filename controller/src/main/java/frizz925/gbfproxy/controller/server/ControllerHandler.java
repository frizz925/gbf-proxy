package frizz925.gbfproxy.controller.server;

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.List;
import java.util.Map;

import org.apache.commons.io.IOUtils;

import frizz925.gbfproxy.utils.Logger;
import io.undertow.server.HttpHandler;
import io.undertow.server.HttpServerExchange;
import io.undertow.util.HeaderValues;
import io.undertow.util.HttpString;

public class ControllerHandler implements HttpHandler {
    @Override
    public void handleRequest(HttpServerExchange exchange) throws Exception {
        Logger.log("[Controller] Processing new request...");
        URL url = new URL("http", "game.granbluefantasy.jp", exchange.getRequestPath());
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();

        prepareRequest(conn, exchange);
        Logger.log("[Controller] Forwarding request...");
        IOUtils.copy(exchange.getInputStream(), conn.getOutputStream());

        prepareResponse(conn, exchange);
        Logger.log("[Controller] Returning response...");
        if (conn.getResponseCode() >= 400) {
            IOUtils.copy(conn.getErrorStream(), exchange.getOutputStream());
        } else {
            IOUtils.copy(conn.getInputStream(), exchange.getOutputStream());
        }
        Logger.log("[Controller] Request processed.");
    }

    protected void prepareRequest(HttpURLConnection conn, HttpServerExchange exchange) throws Exception {
        conn.setDoOutput(true);
        conn.setRequestMethod(exchange.getRequestMethod().toString());
        for (HeaderValues header : exchange.getRequestHeaders()) {
            String key = header.getHeaderName().toString();
            for (String value : header) {
                conn.addRequestProperty(key, value);
            }
        }
    }

    protected void prepareResponse(HttpURLConnection conn, HttpServerExchange exchange) throws IOException {
        exchange.setStatusCode(conn.getResponseCode());
        Map<String, List<String>> headers = conn.getHeaderFields();
        for (Map.Entry<String, List<String>> entry : headers.entrySet()) {
            String key = entry.getKey();
            if (key == null) {
                continue;
            }
            HttpString name = new HttpString(key);
            List<String> values = entry.getValue();
            exchange.getResponseHeaders().addAll(name, values);
        }
    }
}