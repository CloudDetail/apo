FROM node:22.3.0-alpine3.19 as build

WORKDIR /app/

COPY package.json ./

RUN yarn install

COPY . .

ENV NODE_OPTIONS --openssl-legacy-provider
ARG VITE_PUBLIC_POSTHOG_HOST
ARG VITE_PUBLIC_POSTHOG_KEY

RUN yarn build

FROM nginx:alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

