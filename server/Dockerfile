FROM golang:1.21 AS build-stage

WORKDIR /app

COPY ./ ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

# --------------------------------------------------

FROM gcr.io/distroless/static-debian11 AS release-stage

WORKDIR /

COPY --from=build-stage /main /main

ENTRYPOINT ["/main"]