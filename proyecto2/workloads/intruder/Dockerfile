FROM alpine:3.21
RUN addgroup -S intruder && adduser -S -G intruder intruder
COPY intruder.sh /usr/local/bin/intruder.sh
RUN chmod 0755 /usr/local/bin/intruder.sh
USER intruder
ENTRYPOINT ["/usr/local/bin/intruder.sh"]

