FROM golang:1.25.10-alpine AS builder

WORKDIR /bubble

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bubble ./cmd/bubble

FROM alpine:latest

WORKDIR /bubble

COPY --from=builder /bubble/bubble .
COPY --from=builder /bubble/static ./static
COPY --from=builder /bubble/htmlPages ./htmlPages

EXPOSE 443

CMD ["./bubble"]