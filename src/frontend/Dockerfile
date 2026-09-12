# ── Stage 1: Build Frontend ──────────────────────
FROM node:22-alpine AS builder

WORKDIR /app

# Copy dependency manifests first for layer caching
COPY package.json package-lock.json ./
RUN npm ci --prefer-offline --no-audit

# Copy application source code and build production bundle
COPY . .
RUN npm run build

# ── Stage 2: Serve with Nginx ───────────────────
FROM nginx:1.27-alpine

# Install openssl and generate self-signed certificate for HTTPS localhost
RUN apk add --no-cache openssl && \
    mkdir -p /etc/nginx/ssl && \
    openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
      -keyout /etc/nginx/ssl/localhost.key \
      -out /etc/nginx/ssl/localhost.crt \
      -subj "/C=RU/ST=Dev/L=Local/O=GDH/OU=Dev/CN=localhost" && \
    apk del openssl

# Copy built frontend assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy optimized nginx configuration and template
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Expose standard HTTP and HTTPS ports
EXPOSE 80 443

HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget --spider -q http://localhost:80/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
