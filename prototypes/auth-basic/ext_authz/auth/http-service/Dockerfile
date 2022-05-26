FROM docker.io/library/node:16.14.0-alpine3.15

COPY . /app
RUN cd /app/http-service && ls -lah && npm install

EXPOSE 9002

CMD ["node", "/app/http-service/server"]
