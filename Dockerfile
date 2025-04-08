FROM node:22.13.0-alpine AS node_stage

WORKDIR /usr/app

COPY . .

# Install system dependencies
# RUN apk add --no-cache --virtual .gyp python3 make g++

# Install Node dependencies
RUN npm install --frozen-lockfile

# Set NODE_ENV to production
ENV NODE_ENV=production

# Build production ready application
RUN npm run build

FROM nginx:1.27.3-alpine

WORKDIR /usr/share/nginx/html

COPY --from=node_stage /usr/app/dist .

COPY nginx/templates/default.conf.template /etc/nginx/templates/

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
