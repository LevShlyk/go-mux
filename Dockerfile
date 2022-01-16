FROM golang
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN go build  -o ./ozon-exam
CMD [ "./ozon-exam" ]