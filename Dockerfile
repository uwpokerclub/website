FROM node:lts-alpine

WORKDIR /usr/app

# Install build tools for node-gyp
RUN apk add --no-cache --virtual .gyp python make g++

COPY package*.json ./

COPY ./scripts/entrypoint.sh /entrypoint.sh

# Install dependencies
RUN npm install

COPY . .

EXPOSE 5000

ENTRYPOINT ["/entrypoint.sh"]

CMD ["start"]
