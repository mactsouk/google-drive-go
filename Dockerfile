# docker build -t gdrive .
# docker run -it -p 2345:2345 gdrive

FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

RUN mkdir /pro
ADD ./gDriveList.go /pro/
WORKDIR /pro
RUN go get -d -v ./...
RUN go build -o server gDriveList.go

FROM alpine:latest

RUN mkdir /pro
ADD ./credentials.json /pro/
ADD ./token.json /pro/
COPY --from=builder /pro/server /pro/server
EXPOSE 2345
WORKDIR /pro
CMD ["/pro/server"]
