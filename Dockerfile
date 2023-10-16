FROM golang:1.20.10-bullseye AS builder
WORKDIR /builddir
COPY . .
RUN go build -o /builddir/parser ./cmd/parser
RUN go build -o /builddir/config ./cmd/config

FROM ubuntu:jammy
WORKDIR /app
COPY --from=builder /builddir/parser /app/parser
COPY --from=builder /builddir/config /app/config
CMD ["/app/parser"]
