
FROM node:16.16 AS nodeBuild

ENV NODE_ENV=production

WORKDIR /app
COPY web/package.json web/yarn.lock ./

RUN yarn install

COPY web/ .

RUN yarn build


FROM golang:1.19.0-alpine AS goBuild

WORKDIR /app

COPY api/go.mod api/go.sum ./
RUN go mod download && go mod verify

COPY api ./

RUN CGO_ENABLED=0 go build -v -o cms ./cmd
COPY --from=nodeBuild /app/build ./web

EXPOSE 8080

#ENTRYPOINT /app/cms

FROM gcr.io/distroless/static

COPY --from=nodeBuild /app/build /web
COPY --from=goBuild /app/cms /

ENTRYPOINT ["/cms"]


