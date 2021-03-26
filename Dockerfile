FROM node:12.16.2-alpine3.11 AS node_stage

# Set work dir
WORKDIR /usr/app

# Install build tools for node-gyp
RUN apk add --no-cache --virtual .gyp python make g++

COPY . .

# Install all deps
RUN apk add --no-cache --virtual .gyp python make g++ \
    && npm install \
    && npm run build

FROM alpine:latest
COPY --from=node_stage /usr/app/build /usr/api/
