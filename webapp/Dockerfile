### ====================== WEBAPP BUILD IMAGE ======================
FROM node:22.14.0-alpine

WORKDIR /usr/app

# Copy package.json and package-lock.json files over
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy files into container
COPY . .

# Build and run the React app
CMD ["npm", "start"]