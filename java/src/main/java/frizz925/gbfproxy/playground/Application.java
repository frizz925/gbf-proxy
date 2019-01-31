package frizz925.gbfproxy.playground;

import java.io.StringWriter;
import java.net.HttpURLConnection;
import java.net.URL;

import org.apache.commons.io.IOUtils;

import frizz925.gbfproxy.utils.Logger;

public class Application {
    public static void main(String[] args) {
        try {
            new Application().start();
        } catch (Exception e) {
            Logger.error(e);
        }
    }

    public void start() throws Exception {
        URL url = new URL("https://cdn-connect.mobage.jp/jssdk/mobage-menubar.2.4.3.min.js");
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        IOUtils.copy(conn.getInputStream(), System.out);
    }
}