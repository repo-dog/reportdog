# ---- Frontend build stage ----
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ---- Backend build stage ----
FROM golang:1.26.0-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# ---- Final stage ----
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=backend-builder /server /server
COPY --from=frontend-builder /app/frontend/dist /public

ENV PUBLIC_DIR=/public
EXPOSE 8080
CMD ["/server"]
