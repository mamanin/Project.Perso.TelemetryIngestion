# see: https://templ.guide/quick-start/installation#docker

FROM golang:latest AS fetch-stage
COPY go.mod go.sum ./
WORKDIR /app
RUN go mod download

FROM ghcr.io/a-h/templ:latest AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM golang:latest AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/app

FROM gcr.io/distroless/base-debian12 AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
