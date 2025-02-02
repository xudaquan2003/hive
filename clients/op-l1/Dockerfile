FROM ethereum/client-go:latest AS base
# release-1.13 or latest

RUN apk add --no-cache jq yarn git linux-headers gcompat
RUN /usr/local/bin/geth console --exec 'console.log(admin.nodeInfo.name)' --maxpeers=0 --nodiscover --dev 2>/dev/null | head -1 > /version.txt

# Regenerate the L1 genesis file
RUN mkdir /hive
ADD genesis.json /genesis.json

# Inject the startup script
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Helper scripts; called from devnet.go
ADD cat.sh /hive-bin/cat.sh
RUN chmod +x /hive-bin/cat.sh

VOLUME ["/db"]

EXPOSE 8545/tcp
EXPOSE 8546/tcp

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]












