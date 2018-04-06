FROM golang:latest as builder
ADD . /usr/local/go/src/github.com/LINBIT/linstor-external-provisioner/
RUN cd /usr/local/go/src/github.com/LINBIT/linstor-external-provisioner && \
		 make staticrelease && \
		 mv ./linstor-external-provisioner-linux-amd64 /

FROM drbd.io/linstor-client
MAINTAINER Roland Kammerer <roland.kammerer@linbit.com>
COPY --from=builder /linstor-external-provisioner-linux-amd64 /linstor-external-provisioner
ENTRYPOINT ["/linstor-external-provisioner"]
