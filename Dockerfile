FROM scratch
ADD server /
EXPOSE 80
ENTRYPOINT ["/server"]
CMD ["-p", "80"]
