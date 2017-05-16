FROM alpine:3.5

COPY build/cord-network-manager /cord/network-manager

CMD /cord/network-manager