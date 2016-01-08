FROM golang:1.5.1
MAINTAINER Ludovic Post <ludovic.post@epitech.eu>

EXPOSE 3000

COPY . src/github.com/PepperSalt42/api

RUN go get github.com/PepperSalt42/api
RUN go install github.com/PepperSalt42/api

CMD ["api"]
