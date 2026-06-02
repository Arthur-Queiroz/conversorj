FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

FROM caddy:2-alpine
COPY --from=frontend /app/dist /srv
