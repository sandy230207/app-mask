FROM golang:1.15-alpine AS builder
# WORKDIR /usr/local/go/src/app-mask
WORKDIR /app-mask
ENV GO111MODULE on
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./main.go

FROM alpine:latest
COPY --from=builder /app-mask/server .
EXPOSE 3000
ENTRYPOINT ./server