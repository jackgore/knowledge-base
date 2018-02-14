FROM golang:1.9

RUN go get -u github.com/golang/dep/cmd/dep && go install github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/JonathonGore/knowledge-base
COPY . /go/src/github.com/JonathonGore/knowledge-base

RUN dep ensure -v && go install -v

ENTRYPOINT ["go-service"]
