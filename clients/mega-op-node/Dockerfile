FROM megaeth-labs/devnet/op-node:latest AS base

RUN apk add --no-cache jq && apk add --update curl

RUN /usr/local/bin/op-node --version  > /version.txt

# Inject the startup script
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh


ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]