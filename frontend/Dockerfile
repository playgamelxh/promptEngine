# Build stage
FROM docker.m.daocloud.io/library/node:18-alpine AS builder

WORKDIR /app

COPY package.json ./
RUN npm config set registry https://registry.npmmirror.com
# Clean install to ensure packages are correct
RUN rm -rf node_modules package-lock.json
RUN npm install

COPY . .
RUN npm run build

# Production stage
FROM docker.m.daocloud.io/library/nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
