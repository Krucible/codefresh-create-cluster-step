FROM golang:1.14

WORKDIR /home/

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY *.go .
RUN go build .

RUN curl -fL https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectl > /usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl

RUN curl -fL https://github.com/Krucible/krucible-cli/releases/download/v0.1.4/krucible-linux-amd64 > /usr/local/bin/krucible
RUN chmod +x /usr/local/bin/krucible

CMD ["codefresh-create-cluster-step"]
