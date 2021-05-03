FROM uwpokerclub/app:latest AS app

FROM node:14.16.0-alpine3.11

WORKDIR /usr/api

COPY . .

ADD ./.profile.d /app/.profile.d

RUN apk update && apk upgrade && apk add bash curl openssh

RUN apk add --no-cache --virtual python make g++ \
    && npm install \
    && npm run build

COPY --from=app /usr/app/build ./dist/build/

RUN rm /bin/sh && ln -s /bin/bash /bin/sh

CMD npm run start:prod
