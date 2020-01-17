FROM scratch
#COPY ui/build/ /build
#COPY ui/index.html /
#ENV ZONEINFO=/zoneinfo.zip
#COPY zoneinfo.zip /
COPY config-reloader /
ENTRYPOINT ["/config-reloader"]
