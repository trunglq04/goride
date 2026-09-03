FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY build/auth-service /app/build/auth-service
ENTRYPOINT ["/app/build/auth-service"]
