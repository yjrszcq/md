FROM node:20-alpine as nodejs
COPY ./web /build/web
WORKDIR /build/web
RUN npm config set registry https://registry.npmmirror.com
RUN npm install
RUN npm run build

FROM golang:1.22-alpine as go
ENV GOPROXY=https://goproxy.cn,direct
COPY ./md /build/md
COPY --from=nodejs /build/web/dist /build/md/web
WORKDIR /build/md
RUN go build

FROM alpine:latest
COPY --from=go /build/md/md /md/
ENV reg=true
ENV ai_key=md-ai-encrypt-key-2024
EXPOSE 9900
RUN chmod +x /md/md
CMD /md/md -p 9900 -log /md/logs -data /md/data -reg=${reg} -pg_host=${pg_host} -pg_port=${pg_port} -pg_user=${pg_user} -pg_password=${pg_password} -pg_db=${pg_db} -ai_key=${ai_key}
