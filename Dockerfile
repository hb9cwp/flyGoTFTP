# Re-factored TFTP server with pin/tftp library based on the
# Dockerfile of goStatic Web server:
#
# Note: this Dockerfile of goStatic really produces minimal Distroless Container images with very small size:
#  https://github.com/PierreZ/goStatic/blob/master/Dockerfile

FROM scratch
MAINTAINER hb9cwp

# stage 0
FROM --platform=$BUILDPLATFORM golang:latest AS builder

ARG TARGETPLATFORM

# get source from repo:
# https://github.com/PierreZ/goStatic
#WORKDIR /go/src/github.com/PierreZ/goStatic
# https://github.com/pin/tftp
WORKDIR /go/src/github.com/pin/tftp
# add some changes from current local dir, such as tftpServer.go:
COPY . .

RUN mkdir ./bin && \
    apt-get update && apt-get install -y upx && \

    # getting right vars from docker buildx
    # especially to handle linux/arm/v6 for example
    GOOS=$(echo $TARGETPLATFORM | cut -f1 -d/) && \
    GOARCH=$(echo $TARGETPLATFORM | cut -f2 -d/) && \
    GOARM=$(echo $TARGETPLATFORM | cut -f3 -d/ | sed "s/v//" ) && \

    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build ${BUILD_ARGS} -ldflags="-s" -tags netgo -installsuffix netgo -o ./bin/tftpServer && \

    mkdir ./bin/etc && \
    ID=$(shuf -i 100-9999 -n 1) && \
    upx -9 ./bin/tftpServer && \
    echo $ID && \
    echo "appuser:x:$ID:$ID::/sbin/nologin:/bin/false" > ./bin/etc/passwd && \
    echo "appgroup:x:$ID:appuser" > ./bin/etc/group

# stage 1
FROM scratch
WORKDIR /
COPY --from=builder /go/src/github.com/pin/tftp/bin/ .

# tftpServer enforces /app/tftp/ as path for files it reads/serves
# to prevent any dir tree traversal attacks, such as /etc/hosts, ../../b
# Note: mkdir is not required (it will fail in scratch!), because COPY below
# creates /app/tftp/
#RUN mkdir -p /app/tftp

# put some content for TFTP to serve
COPY --from=builder /go/src/github.com/pin/tftp/go.mod /app/tftp/
# add more files...

# must run as 'root' if tftpServer opens low port 69/udp, otherwise
# 'appuser' is sufficient to open high port such as 6969/udp and port fwd in fly.toml
USER root
#USER appuser

ENTRYPOINT ["/tftpServer"]
#CMD ["arg1", "arg2"]
