# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 2: Build Backend
FROM golang:alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum* ./
# If go.sum doesn't exist yet, it's fine. We will generate it.
RUN go mod download || true 
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
RUN go build -o wwchan-server ./cmd/main.go

# Stage 3: Final Image
FROM alpine:latest
WORKDIR /app
COPY --from=backend-builder /app/wwchan-server .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Create uploads directory
RUN mkdir -p /app/uploads

EXPOSE 8080
CMD ["./wwchan-server"]
