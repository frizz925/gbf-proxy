package frizz925.gbfproxy.proxy.channels;

import java.io.IOException;
import java.nio.channels.SelectionKey;

public interface WriteHandler extends ChannelHandler {
    void write(SelectionKey key) throws IOException;
}