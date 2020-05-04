FROM golang:1.14

WORKDIR /home/

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build .

CMD ["codefresh-create-cluster-step"]
