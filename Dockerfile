# build stage

# docker build . -t rewanthtammana/grumpy:nonscratch
# docker push rewanthtammana/grumpy:nonscratch

FROM golang:1.10-stretch AS build-env
RUN mkdir -p /go/src/github.com/testinguser883/grumpy
WORKDIR /go/src/github.com/testinguser883/grumpy
COPY  . .
RUN useradd -u 10001 webhook
RUN go get -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o grumpywebhook

FROM scratch
COPY --from=build-env /go/src/github.com/testinguser883/grumpy/grumpywebhook .
COPY --from=build-env /go/src/github.com/testinguser883/grumpy/notary-slim .
COPY --from=build-env /etc/passwd /etc/passwd
USER webhook
ENTRYPOINT ["/grumpywebhook"]
