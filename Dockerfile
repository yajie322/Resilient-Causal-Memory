FROM golang:latest

RUN mkdir /svr
WORKDIR /svr
COPY . .
RUN go build -o main .

CMD ["/svr/main"]