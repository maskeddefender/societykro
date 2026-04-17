FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go-common dependency first (for layer caching)
COPY packages/go-common/ /packages/go-common/

# Copy auth-service
COPY services/auth-service/go.mod services/auth-service/go.sum ./
RUN go mod download

COPY services/auth-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/auth-service cmd/server/main.go

# Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/auth-service /bin/auth-service
EXPOSE 8081
CMD ["/bin/auth-service"]
