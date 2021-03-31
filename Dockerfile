FROM uwpokerclub/app:latest AS app

FROM node:14.16.0-alpine3.11

WORKDIR /usr/api

COPY . .

RUN apk add --no-cache --virtual .gyp python make g++ \
    && npm install \
    && npm run build

COPY --from=app /usr/app/build ./dist/build/

CMD npm run migrate up \
    && npm run start:prod
