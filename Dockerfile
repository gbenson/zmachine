ARG BUILDER_IMAGE=scratch

FROM ${BUILDER_IMAGE}
ARG BUILDER_UID=error
ARG BUILDER_GID=error

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get -y update
RUN apt-get -y install --no-install-recommends \
      bsdextrautils \
      less \
      libsdl2-dev \
      nano \
      patch \
      sudo

RUN addgroup --gid ${BUILDER_GID} group
RUN adduser --uid ${BUILDER_UID} --gid ${BUILDER_GID} user
RUN echo 'user ALL=(ALL) NOPASSWD: ALL' >>/etc/sudoers

WORKDIR /home/user

RUN install -d -ouser -ggroup -m700 \
      /home/user/go \
      /home/user/.cache/go-build

VOLUME /home/user/go
VOLUME /home/user/.cache/go-build

USER user:group
