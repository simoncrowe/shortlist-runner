FROM golang:1.22

WORKDIR /src

COPY go.mod go.sum  ./
COPY internal ./internal
COPY cmd ./cmd

RUN go build  -o /opt/reticle/runner ./cmd/runner/

CMD ["/opt/reticle/runner"]
