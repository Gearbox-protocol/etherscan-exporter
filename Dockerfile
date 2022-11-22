FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY . /app/

ENTRYPOINT ["/app/exporter"]
