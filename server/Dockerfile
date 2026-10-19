# 아지트 서버 이미지.
# server 가 ../proto 를 replace 로 쓰기 때문에, 빌드는 저장소 루트에서 한다:
#   docker build -f server/Dockerfile -t azit-server .
# 운영 EC2 가 ARM(t4g) 이라 Actions 에서는 --platform linux/arm64 로 빌드한다.

# 1단계: Go 로 바이너리 만들기 (Spring 의 gradle bootJar 단계)
# 빌드는 빌드하는 컴퓨터의 CPU 에서 하고(BUILDPLATFORM),
# GOARCH 만 목표 CPU(TARGETARCH)로 바꿔서 크로스 컴파일한다. 에뮬레이션이 없어 빠르다.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

# go.work 에는 desktop 도 들어 있어서, 서버만 빌드할 땐 workspace 를 끈다.
# glebarez/sqlite 는 순수 Go 라 CGO 없이 빌드된다.
ENV GOWORK=off CGO_ENABLED=0

# 의존성만 먼저 받아 두면, 코드만 바뀐 빌드에서는 이 단계가 캐시로 재사용된다.
COPY proto/ proto/
COPY server/go.mod server/go.sum server/
RUN cd server && go mod download

COPY server/ server/
RUN cd server && GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/azit-server .

# 2단계: 바이너리만 담은 작은 실행 이미지 (JRE 이미지에 jar 하나 넣는 것과 비슷)
FROM alpine:3.24

# GHCR 에서 이 이미지가 저장소에 연결되도록 (Actions 토큰으로 push/pull 하려면 필요)
LABEL org.opencontainers.image.source=https://github.com/TogetherBros/azit

# ca-certificates: 구글 로그인, R2 에 HTTPS 로 붙을 때 필요
# /data: SQLite 파일이 사는 곳 (compose 에서 볼륨으로 붙인다)
RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -u 10001 azit \
	&& mkdir /data && chown azit:azit /data

COPY --from=build /out/azit-server /usr/local/bin/azit-server

USER azit
WORKDIR /data
ENV PORT=8081 DB_DSN=/data/azit.db
EXPOSE 8081

ENTRYPOINT ["azit-server"]
