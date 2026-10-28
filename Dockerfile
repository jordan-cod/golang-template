ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /src
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

FROM base AS dev
ARG AIR_VERSION=v1.63.4
RUN go install github.com/air-verse/air@${AIR_VERSION}
CMD ["air", "-c", ".air.toml"]

FROM base AS build
ARG VERSION=dev
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/app ./cmd/app

FROM gcr.io/distroless/static-debian13:nonroot AS prod
COPY --from=build /out/app /app
ENV APP_ENV=production
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
