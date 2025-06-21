FROM node:22-alpine AS build
WORKDIR /app
COPY blessed/package.json blessed/package-lock.json ./
RUN npm ci
COPY blessed/ ./
RUN npm run build

FROM nginx:1.27.0-alpine

# Copy built React app
COPY --from=build /app/build /usr/share/nginx/html

# Copy Nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf
COPY blessed.conf /etc/nginx/sites-available/blessed.conf

# Create symbolic link to enable the site
RUN mkdir -p /etc/nginx/sites-enabled && \
    ln -s /etc/nginx/sites-available/blessed.conf /etc/nginx/sites-enabled/

# Copy SSL certificates
COPY ssl/certificate.crt /etc/ssl/certs/certificate.crt
COPY ssl/private.key /etc/ssl/private/private.key

# Expose ports
EXPOSE 80
EXPOSE 443

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
