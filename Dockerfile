FROM node:16.14.2-alpine3.14

WORKDIR /usr/api

COPY . .

ADD ./.profile.d /app/.profile.d

RUN apk update && apk upgrade && apk add bash curl openssh

RUN apk add --no-cache --virtual python make g++

RUN npm install

RUN npm run build

RUN rm /bin/sh && ln -s /bin/bash /bin/sh

RUN mkdir postgres-ca

CMD npm run start:prod
