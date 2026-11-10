FROM alpine:3.20
RUN apk add --no-cache g++ openjdk17-jdk python3 && adduser -D -u 10001 judge
USER judge
WORKDIR /work