### ====================== WEBAPP BUILD IMAGE ======================
FROM node:24.0.1-alpine

WORKDIR /usr/app

# Build argument for GitHub Package authentication
ARG GITHUB_PACKAGE_TOKEN

# Copy package.json, package-lock.json, and .npmrc files over
COPY package.json package-lock.json .npmrc ./

# Install dependencies
RUN npm ci

# Copy files into container
COPY . .

# Build and run the React app
CMD ["npm", "start"]