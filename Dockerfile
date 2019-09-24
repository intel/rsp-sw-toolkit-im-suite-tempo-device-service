FROM scratch
ADD cmd /
EXPOSE 49993
ENTRYPOINT ["/tempo-device-service"]
CMD ["--registry=consul://edgex-core-consul:8500","--profile=docker","--confdir=/res"]
