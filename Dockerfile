# syntax=docker/dockerfile:1

FROM golang:1.20 AS build

COPY ../ /srv/

WORKDIR /srv

RUN go build -o ./bin/app ./main.go

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /srv/bin/app /app

USER nonroot:nonroot

ENTRYPOINT ["/app"]
