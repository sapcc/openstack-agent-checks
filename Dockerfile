FROM golang:1.19.5-alpine3.17 as builder

RUN apk add --no-cache gcc git make musl-dev bash

COPY . /src
RUN make -C /src bin

FROM alpine:3.17
COPY --from=builder /src/bin /usr/bin/

LABEL source_repository="https://github.com/sapcc/openstack-agent-checks"
USER nobody:nobody
CMD ["/usr/bin/neutron-linuxbridge-dhcp-exporter"]
