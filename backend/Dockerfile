FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o conversorj .

FROM alpine:3.20
RUN apk add --no-cache ffmpeg python3 curl ca-certificates tzdata
RUN curl -fsSL https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
        -o /usr/local/bin/yt-dlp \
    && chmod +x /usr/local/bin/yt-dlp
WORKDIR /app
COPY --from=builder /app/conversorj .
RUN mkdir -p data tmp
ENV DB_PATH=/app/data/conversorj.db \
    TMP_DIR=/app/tmp
EXPOSE 8080
CMD ["./conversorj"]
