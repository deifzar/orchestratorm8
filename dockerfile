# Build
FROM golang:1.24.0-alpine AS build-env
RUN apk add build-base
WORKDIR /app
COPY . /app
RUN go mod download
RUN go build .

# Release
FROM alpine:3.20.3
COPY --from=build-env /app/orchestratorm8 /usr/local/bin/
CMD ["orchestratorm8","--help"]