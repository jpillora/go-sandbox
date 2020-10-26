FROM golang:buster
WORKDIR /go/src/app
COPY . .
RUN go build -v -o /bin/app
