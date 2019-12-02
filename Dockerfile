FROM ubuntu:18.04
COPY bin/avatar-fight-server /usr/bin
CMD ["/usr/bin/avatar-fight-server","-goserver","./goserver.json"]
