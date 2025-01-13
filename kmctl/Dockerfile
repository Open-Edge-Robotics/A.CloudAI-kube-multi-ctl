FROM docker.io/golang:1.23.3-alpine AS builder

WORKDIR /app

COPY . .

ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl-dev

RUN go mod download

RUN go build -o app server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/app .

EXPOSE 50051

ENTRYPOINT [ "/bin/sh", "-c" ]

CMD ["./app serve"]