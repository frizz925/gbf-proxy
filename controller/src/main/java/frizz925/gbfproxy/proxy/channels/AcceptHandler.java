package frizz925.gbfproxy.proxy.channels;

import java.io.IOException;
import java.nio.channels.SelectionKey;

public interface AcceptHandler extends ChannelHandler {
    void accept(SelectionKey key) throws IOException;
}