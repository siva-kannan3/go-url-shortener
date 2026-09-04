FROM golang:1.26 AS builder

COPY . .

RUN go mod download

RUN GOOS=linux CGO_ENABLED=0 go build -o /build

FROM alpine:latest AS runner

COPY --from=builder /build /app/build 

EXPOSE 8000

CMD ["/app/build"]