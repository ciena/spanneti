FROM alpine:3.5

COPY build/spanneti /cord/spanneti

CMD /cord/spanneti