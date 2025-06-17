
FROM node:22-alpine AS build
WORKDIR /app
COPY blessed/package.json blessed/package-lock.json ./
RUN npm ci
COPY blessed/ ./
RUN npm run build

FROM nginx:1.27.0-alpine

RUN rm -f /etc/nginx/conf.d/default.conf

COPY --from=build /app/build /usr/share/nginx/html

COPY nginx.conf /etc/nginx/nginx.conf

COPY blessed.conf /etc/nginx/conf.d/blessed.conf

COPY ssl/certificate.crt /etc/ssl/certs/certificate.crt
COPY ssl/private.key /etc/ssl/private/private.key

EXPOSE 80
EXPOSE 443

CMD ["nginx", "-g", "daemon off;"]
