FROM docker.io/library/node:16.14.0-alpine3.15

COPY . /app
WORKDIR /app/http-service
RUN npm install

EXPOSE 9002

CMD ["node", "/app/http-service/server"]
