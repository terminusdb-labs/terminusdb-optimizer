FROM golang:1.16-alpine

WORKDIR /app

COPY . /app

RUN go build -o listener main.go

EXPOSE 1053

CMD [ "/app/listener" ]
