FROM debian:bullseye
WORKDIR /
COPY build/localnet/start.sh /start.sh
COPY build/localnet/restart.sh /restart.sh
COPY bin/pocket-linux /usr/local/bin/pocket
CMD ["/usr/local/bin/pocket"]
ENTRYPOINT ["/start.sh", "/usr/local/bin/pocket"]
