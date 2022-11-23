FROM golang:1.16-alpine as builder
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 go build -o listener main.go

FROM scratch
WORKDIR /app
COPY --from=builder /app/listener /app/listener
ENTRYPOINT [ "/app/listener" ]
