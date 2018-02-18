FROM golang:1.9.4-alpine3.7
RUN apk add --no-cache make
WORKDIR /go/src/github.com/sgreben/tj/
COPY . .
RUN make binaries/linux_x86_64/tj

FROM scratch
COPY --from=0 /go/src/github.com/sgreben/tj/binaries/linux_x86_64/tj /tj
ENTRYPOINT [ "/tj" ]