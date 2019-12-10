# Apache v2 license
# Copyright (C) 2019 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# ---------------------------------------------------------
FROM golang:1.12 as builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN make -e SERVICE_NAME=service build && mkdir logs

# ---------------------------------------------------------
FROM scratch as service
ARG APP_PORT=49993

# ARG variable substitution doesn't work with --chown below 19.03.0 :(
# https://github.com/moby/moby/issues/35018
COPY --from=builder --chown=2000:2000 /app/logs /logs
COPY --from=builder --chown=2000:2000 /app/service /
COPY cmd/res /res
COPY LICENSE-APACHE-2.0.txt .

USER 2000
ENV APP_PORT=$APP_PORT
EXPOSE $APP_PORT 9001
ENTRYPOINT ["/service"]
CMD ["--registry=consul://edgex-core-consul:8500", "--profile=docker","--confdir=/res"]
