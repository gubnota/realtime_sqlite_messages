FROM golang:1.24-alpine AS builder

# Add required build tools for CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY .env .env
RUN CGO_ENABLED=1 GOOS=linux go build -o /realtime_sqlite_messages

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /
COPY --from=builder /realtime_sqlite_messages /realtime_sqlite_messages
COPY --from=builder /app/.env /.env

#ENV SQLITE_PATH=/data/auth.db
#ENV JWT_SECRET=your-256-bit-secret
#ENV PORT=8080

VOLUME /data
EXPOSE 8080
CMD ["/realtime_sqlite_messages"]