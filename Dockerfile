FROM golang

ENV id=0

RUN mkdir /svr
WORKDIR /svr
COPY /svr /svr
RUN go build -o main .

CMD ["/svr/main"]