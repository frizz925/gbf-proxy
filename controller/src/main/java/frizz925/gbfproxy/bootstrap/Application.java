package frizz925.gbfproxy.bootstrap;

import frizz925.gbfproxy.bootstrap.threads.ControllerThread;
import frizz925.gbfproxy.bootstrap.threads.ProxyThread;
import frizz925.gbfproxy.bootstrap.threads.ServerThread;
import frizz925.gbfproxy.utils.Logger;

class Application {
    public static void main(String[] args) {
        new Application().start();
    }

    public void start() {
        startController();
        startProxy();
    }

    public void startController() {
        ServerThread controller = new ControllerThread(8000);
        Logger.log("Starting controller server...");
        controller.start();
        Logger.log("Controller server listening at :" + controller.port);
    }

    public void startProxy() {
        ServerThread proxy = new ProxyThread(8080);
        Logger.log("Starting proxy server...");
        proxy.start();
        Logger.log("Proxy server listening at :" + proxy.port);
    }
}