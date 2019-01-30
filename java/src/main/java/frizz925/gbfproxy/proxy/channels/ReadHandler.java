package frizz925.gbfproxy.proxy.channels;

import java.io.IOException;
import java.nio.channels.SelectionKey;

public interface ReadHandler extends ChannelHandler {
    void read(SelectionKey key) throws IOException;
}