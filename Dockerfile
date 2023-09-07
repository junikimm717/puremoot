FROM golang:latest as bot

WORKDIR /build
COPY go.mod go.sum /build/
RUN go mod download
COPY . /build/
RUN go build

FROM redis:latest
WORKDIR /data
RUN apt-get update && apt-get install -y ca-certificates procps && apt-get clean
COPY --from=bot /build/puremoot /puremoot
COPY entrypoint.sh /entrypoint.sh

CMD ["/entrypoint.sh"]
