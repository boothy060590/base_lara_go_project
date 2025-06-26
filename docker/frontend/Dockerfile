FROM node:20-alpine as dev
RUN npm install -g npm@9
RUN echo "Node version is:" && node -v

# install simple http server for serving static content
RUN apk update \
    && apk add --update bash util-linux make g++ \
    && rm -rf /var/cache/apk/*

# make the 'app' folder the current working directory
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install

COPY . .

RUN ls -l /app/node_modules

EXPOSE 5173
ENV PATH="./node_modules/.bin:$PATH"
CMD [ "npx", "vite" ] 