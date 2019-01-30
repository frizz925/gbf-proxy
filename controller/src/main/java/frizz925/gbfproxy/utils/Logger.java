package frizz925.gbfproxy.utils;

public class Logger {
    public static void log(String message) {
        System.out.println(message);
    }

    public static void error(Exception e) {
        e.printStackTrace();
    }
}