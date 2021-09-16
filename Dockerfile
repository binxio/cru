FROM 		golang:1.17

WORKDIR		/cru
ADD		. /cru
RUN		CGO_ENABLED=0 GOOS=linux go build  -ldflags '-extldflags "-static"' .

FROM 		scratch
COPY --from=0		/cru/cru /
ENTRYPOINT 	["/cru"]
