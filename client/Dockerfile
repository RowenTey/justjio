# build phase
FROM node:hydrogen-alpine AS build-env

WORKDIR /app
COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# runtime phase
FROM nginx:1.25.0-alpine AS production
RUN rm -rf /usr/share/nginx/html/*

COPY ${NGINX_CONFIG_FILE:-nginx.conf} /etc/nginx/conf.d/default.conf
COPY --from=build-env app/dist /usr/share/nginx/html/web-app
COPY --from=build-env app/dist/* /tmpl/dist/web-app/

COPY env.sh /docker-entrypoint.d/env.sh
RUN dos2unix /docker-entrypoint.d/env.sh
RUN chmod +x /docker-entrypoint.d/env.sh

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]