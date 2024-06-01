FROM golang:1.22 AS build
ARG DEBUG=false
ARG DOMAIN=localhost
ARG HTTP_PORT=8000
ARG HTTPS_PORT=4430
WORKDIR /usr/src/app
RUN apt-get update && apt-get -y install nodejs npm libcairo-dev rsync
RUN npm i -g google-closure-compiler
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN echo ${DEBUG} > DEBUG
RUN echo ${DOMAIN} > DOMAIN
RUN echo ${HTTP_PORT} > HTTP_PORT
RUN echo ${HTTPS_PORT} > HTTPS_PORT
RUN make
RUN openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout build/key -out build/cert -subj "/C=US/ST=XXX/L=XXX/O=XXX/CN=www.example.com"

FROM debian:bookworm
RUN apt-get update && apt-get -y install libcairo2
COPY --from=build /usr/src/app/build /usr/src/app/build
WORKDIR /usr/src/app/build
CMD exec ./vandal -f -p $HTTP_PORT -sp $HTTPS_PORT -host $DOMAIN
