FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/main .

FROM scratch
LABEL org.opencontainers.image.source=https://github.com/clementd64/k8s-example-app
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
