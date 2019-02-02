FROM golang:1.10
COPY ./bin/gbf-proxy-linux-amd64 $GOPATH/bin/gbf-proxy
ENV GBF_PROXY_PORT 80
CMD gbf-proxy ${GBF_PROXY_PORT}
