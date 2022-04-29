FROM golang:1.17.5-bullseye
MAINTAINER jknedlik <j.knedlik@gsi.de>
RUN apt-get update
COPY . /go/lustre_exporter
WORKDIR /go/lustre_exporter
#build lustre_exporter
RUN make build
#rename /move binary
RUN mkdir /build
RUN mv lustre_exporter /build/lustre_exporter-$(cat VERSION)
#prepare mountpoint for copy
RUN mkdir /cpy
ENTRYPOINT ["/bin/bash"]
ENTRYPOINT ["cp"]
CMD ["-r","/build","/cpy/"]
