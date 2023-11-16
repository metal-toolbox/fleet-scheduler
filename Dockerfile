FROM alpine:latest

ENTRYPOINT ["/usr/sbin/fleet-scheduler"]

COPY fleet-scheduler /usr/sbin/fleet-scheduler
RUN chmod +x /usr/sbin/fleet-scheduler