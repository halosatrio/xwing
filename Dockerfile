FROM golang:1.22.1-alpine AS build

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

WORKDIR /dist

RUN cp /build/main .
RUN cp /build/.env .

FROM scratch AS runtime

WORKDIR /app

COPY --from=build /dist/main .
COPY --from=build /build/.env .

CMD [ "./main" ]