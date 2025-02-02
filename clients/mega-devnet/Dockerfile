FROM megaeth-labs/testnet-all/op-reth:latest AS base

RUN apt-get update && apt-get install -y jq

# Inject the startup script
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Create version.txt
# RUN /usr/local/bin/devnet --version > /version.txt
RUN echo "v0.0.0" > /version.txt

VOLUME ["/db"]

EXPOSE 8545/tcp
EXPOSE 8546/tcp
EXPOSE 8551/tcp

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]

