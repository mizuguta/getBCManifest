FROM golang
RUN go get github.com/joho/godotenv
RUN go get github.com/gorilla/mux

COPY app /go/src/app
WORKDIR /go/src/app

RUN go build main.go getManifestInfo.go ManifestInfo.go
ENTRYPOINT /go/src/app/main

EXPOSE 8081