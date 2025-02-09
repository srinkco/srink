# Build Stage: Build bot using the alpine image, also install doppler in it
FROM golang:1.20-alpine AS builder
RUN apk add --no-cache curl wget gnupg git upx
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=`go env GOHOSTOS` GOARCH=`go env GOHOSTARCH` go build -o out/srink -ldflags="-w -s" .
# RUN upx --brute out/srink
RUN git clone https://github.com/srinkco/website

# Run Stage: Run bot using the bot binary copied from build stage
FROM alpine:3.18
COPY --from=builder /app/out/srink /app/srink
COPY --from=builder /app/website /app/frontend
CMD ["/app/srink", "server"]
