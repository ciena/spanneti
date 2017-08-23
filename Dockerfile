FROM alpine:3.5

COPY build/spanneti /bin/spanneti

CMD ["spanneti"]
