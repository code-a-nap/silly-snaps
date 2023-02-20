# syntax=docker/dockerfile:1

FROM golang:latest


RUN mkdir /app
RUN mkdir /app/logs
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY *.html ./
COPY *.js ./
COPY static ./static

RUN go build -o silly-snaps


FROM node:16

RUN apt-get update && apt-get install -y wget gnupg && wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list' && apt-get update  && apt-get install -y google-chrome-stable fonts-ipafont-gothic fonts-wqy-zenhei fonts-thai-tlwg fonts-kacst fonts-freefont-ttf libxss1 libgles2 libegl1 --no-install-recommends && rm -rf /var/lib/apt/lists/*

RUN mkdir /app
RUN mkdir /app/logs

COPY *.html ./app
COPY *.js ./app
COPY static ./app/static
COPY --from=0 /app/silly-snaps /app

WORKDIR /app
RUN npm init -y && npm i puppeteer && npm install jsonwebtoken

EXPOSE 80

CMD [ "./silly-snaps" ]