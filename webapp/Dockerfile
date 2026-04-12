### ====================== WEBAPP BUILD IMAGE ======================
FROM node:24.14.1-alpine

WORKDIR /usr/app

# Copy package.json, package-lock.json, and .npmrc files over
COPY package.json package-lock.json .npmrc ./

# Install dependencies using secret mount for GitHub token
RUN --mount=type=secret,id=GITHUB_PACKAGE_TOKEN \
    GITHUB_PACKAGE_TOKEN=$(cat /run/secrets/GITHUB_PACKAGE_TOKEN) npm ci

# Copy files into container
COPY . .

# Build and run the React app
CMD ["npm", "start"]