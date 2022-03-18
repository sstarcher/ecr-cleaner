FROM --platform=$BUILDPLATFORM golang:1.14 as builder

WORKDIR /go/src/github.com/sstarcher/ecr-cleaner
COPY . /go/src/github.com/sstarcher/ecr-cleaner

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /go/bin/ecr-cleaner /go/src/github.com/sstarcher/ecr-cleaner/main.go

FROM alpine:3
RUN apk --update add ca-certificates
RUN addgroup -S ecr-cleaner && adduser -S -G ecr-cleaner ecr-cleaner
USER ecr-cleaner
COPY --from=builder /go/bin/ecr-cleaner /usr/local/bin/ecr-cleaner

ENTRYPOINT ["ecr-cleaner"]
