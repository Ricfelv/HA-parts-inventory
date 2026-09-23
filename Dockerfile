FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/parts-inventory ./cmd/parts-inventory

FROM alpine:3.22
ARG BUILD_VERSION
ARG BUILD_ARCH
LABEL io.hass.version="${BUILD_VERSION}" \
      io.hass.type="app" \
      io.hass.arch="${BUILD_ARCH}"
COPY --from=build /out/parts-inventory /usr/local/bin/parts-inventory
ENV PARTS_LISTEN_ADDR=:8080
ENV PARTS_DB_PATH=/data/parts.db
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/parts-inventory"]
