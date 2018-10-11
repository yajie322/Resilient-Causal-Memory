FROM golang

ENV id=0
ENV type=s

RUN mkdir /svr
WORKDIR /svr
COPY /svr /svr
RUN go build -o main .

CMD ["/svr/main"]
