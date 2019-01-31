package frizz925.gbfproxy.bootstrap;

import frizz925.gbfproxy.bootstrap.threads.ControllerThread;
import frizz925.gbfproxy.bootstrap.threads.ProxyThread;
import frizz925.gbfproxy.bootstrap.threads.ServerThread;
import frizz925.gbfproxy.utils.Logger;

public class Application {
    public static final String APP_NAME = "Granblue Proxy";
    public static final String APP_VERSION = "0.1-alpha";

    public static final int CONTROLLER_PORT = 8000;
    public static final int PROXY_PORT = 8088;

    public static String getFullName() {
        return APP_NAME + " " + APP_VERSION;
    }

    public static void main(String[] args) {
        new Application().start();
    }

    public void start() {
        try {
            Thread controller = startController();
            Thread proxy = startProxy((ServerThread) controller);
            controller.join();
            proxy.join();
        } catch (InterruptedException e) {
            // do nothing
        }
    }

    public ServerThread startController() {
        ServerThread controller = new ControllerThread(8000);
        Logger.log("Starting controller server...");
        controller.start();
        Logger.log("Controller server listening at :" + controller.port);
        return controller;
    }

    public ServerThread startProxy(ServerThread backend) {
        ServerThread proxy = new ProxyThread(8088, backend.port);
        Logger.log("Starting proxy server...");
        proxy.start();
        Logger.log("Proxy server listening at :" + proxy.port);
        return proxy;
    }
}