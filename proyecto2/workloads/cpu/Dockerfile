FROM alpine:3.21
RUN apk add --no-cache bc
USER nobody
CMD ["sh", "-c", "timeout 240 sh -c 'while true; do echo 2^20 | bc >/dev/null; done'"]

