FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP_NAME
RUN go build -o /bin/nebula-app ./cmd/${APP_NAME}/main.go

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/nebula-app .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/cmd/gateway/index.html ./cmd/gateway/index.html

EXPOSE 3000 9090

CMD ["./nebula-app"]