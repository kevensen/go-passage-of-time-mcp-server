FROM golang:1.24.5-alpine3.22 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/app go-potms/main.go

FROM golang:1.24.5-alpine3.22
EXPOSE 8080
COPY --from=builder /usr/local/bin/app /usr/local/bin/app
# create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown appuser:appgroup /usr/local/bin/app && \
  chmod 755 /usr/local/bin/app

USER appuser

CMD ["app", "-port", "8080"]