FROM golang:1.20.10-bullseye AS builder
COPY . .
RUN go build -o /builddir/parser ./cmd/parser

FROM ubuntu:jammy
WORKDIR /app
COPY --from=builder /builddir/parser /app/parser
CMD ["/app/parser"]
