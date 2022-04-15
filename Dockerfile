FROM golang:1.17.8-alpine3.15 as builder

RUN apk add libgit2-dev git build-base

COPY . /workdir
WORKDIR /workdir

RUN make build

FROM alpine:3.15.0
WORKDIR /

ADD https://dl.k8s.io/v1.19.16/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN echo -n "6b9d9315877c624097630ac3c9a13f1f7603be39764001da7a080162f85cbc7e  /usr/local/bin/kubectl" | sha256sum -c
RUN chmod +x /usr/local/bin/kubectl

RUN apk add --no-cache git expat libgit2

COPY --from=builder /workdir/templates /templates
COPY --from=builder /workdir/static /root/static
COPY --from=builder /workdir/kube-applier /usr/local/bin/kube-applier

ENTRYPOINT /usr/local/bin/kube-applier
