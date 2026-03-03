# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Serve stage
FROM node:20-alpine
RUN npm install -g serve
COPY --from=builder /app/dist /app/dist
EXPOSE 3000
CMD ["serve", "-s", "/app/dist", "-l", "3000"]
