FROM golang:1.20.10-bullseye AS builder
WORKDIR /builddir
COPY . .
RUN go build -o parser ./cmd/parser

FROM ubuntu:jammy
WORKDIR /app
COPY --from=builder /builddir/parser /app/parser
CMD ["./parser"]
