FROM golang:latest as builder
WORKDIR /temp
ADD ./ .
RUN  go env -w GOPROXY=https://goproxy.cn,direct \
     && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gocached .
FROM scratch
LABEL author="Truman.P.Du"
ENV PROJECT_BASE_DIR /opt/app/gocache
WORKDIR ${PROJECT_BASE_DIR}
COPY --from=builder  /temp/gocached .
EXPOSE 6379
CMD ["/opt/app/gocache/gocached"]
