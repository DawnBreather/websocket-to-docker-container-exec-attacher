FROM debian:latest

COPY ./entrypoints/debian.entrypoint.sh /debian.entrypoint.sh
ENTRYPOINT ["/debian.entrypoint.sh"]
