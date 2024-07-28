# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder
WORKDIR /src
COPY go.mod go.sum  ./
COPY internal ./internal
COPY cmd ./cmd
RUN go build  -o /opt/shortlist/runner ./cmd/runner/

FROM debian:stable-slim
COPY --from=builder /opt/shortlist/runner /opt/shortlist/runner
CMD ["/opt/shortlist/runner"]
