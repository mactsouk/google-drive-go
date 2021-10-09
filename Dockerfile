# docker build -t pro-gmail .
# docker run -it -p 2345:2345 pro-gmail

FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

RUN mkdir /pro
ADD ./gDriveList.go /pro/
WORKDIR /pro
RUN go mod init
RUN go mod tidy
RUN go build -o server gDriveList.go

FROM alpine:latest

RUN mkdir /pro
ADD ./credentials.json /pro/
ADD ./token.json /pro/
COPY --from=builder /pro/server /pro/server
EXPOSE 2349
WORKDIR /pro
CMD ["/pro/server"]
