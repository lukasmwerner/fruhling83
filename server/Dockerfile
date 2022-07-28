FROM golang:alpine AS build

WORKDIR /src/

COPY go.* /src/
RUN go mod download -x

COPY . /src

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build  -o /out/app .

FROM scratch as run

COPY --from=build /out/app /

EXPOSE 8080

ENTRYPOINT ["/app"]
