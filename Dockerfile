FROM gcr.io/distroless/static-debian11:nonroot
COPY config-reloader /
USER nonroot
ENTRYPOINT ["/config-reloader"]
