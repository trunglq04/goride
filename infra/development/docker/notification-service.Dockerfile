FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY build/notification-service /app/build/notification-service
ENTRYPOINT ["/app/build/notification-service"]
