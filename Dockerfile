FROM alpine:3 as ca
RUN apk add --no-cache ca-certificates

FROM 		golang:1.17 as cru

WORKDIR		/cru
ADD		. /cru
RUN		CGO_ENABLED=0 GOOS=linux go build  -ldflags '-extldflags "-static"' .

FROM 		scratch
COPY --from=ca /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=cru		/cru/cru /
ENTRYPOINT 	["/cru"]
