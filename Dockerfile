FROM golang:latest as backend
WORKDIR /build
COPY go.mod go.sum /build/
RUN go mod download
COPY . /build/
RUN go build
CMD ["./puremoot"]
