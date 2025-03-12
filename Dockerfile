FROM golang:1 AS deps
LABEL maintainer='Michaël "e7d" Ferrand <michael@e7d.io>'
WORKDIR /go
EXPOSE 9100

FROM deps AS dev
CMD [ "go", "run", "app.go" ]

FROM deps AS build
COPY src/ /go/src/
WORKDIR /go/src
RUN go get
RUN go build -o /go/app

FROM deps AS prod
COPY --from=build /go/app /go/app
ADD src/gpon_status.lua /go/gpon_status.lua
CMD [ "/go/app" ]
