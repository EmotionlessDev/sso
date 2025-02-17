
FROM golang:1.23.6

WORKDIR /sso

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /sso/bin/main ./cmd/sso/

RUN chmod +x /sso/bin/main

CMD ["/sso/bin/main"]
