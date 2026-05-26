# syntax=docker/dockerfile:1
# auth-ms image. SQLite is pure-Go (modernc), so CGO is off and the binary is
# fully static. The database file lives on a mounted volume at /data.

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY auth-ms/ ./auth-ms/
RUN CGO_ENABLED=0 go build -trimpath -o /out/auth-ms ./auth-ms

FROM alpine:3.20
RUN adduser -D -u 10001 app && mkdir -p /data && chown app:app /data
USER app
COPY --from=build /out/auth-ms /usr/local/bin/auth-ms
ENV AUTH_PORT=9090 AUTH_DB_PATH=/data/auth.db
EXPOSE 9090
VOLUME ["/data"]
ENTRYPOINT ["auth-ms"]
