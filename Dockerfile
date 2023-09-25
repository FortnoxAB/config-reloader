FROM gcr.io/distroless/static-debian12:nonroot
COPY config-reloader /
USER nonroot
ENTRYPOINT ["/config-reloader"]
