FROM scratch
COPY go-aqaramqtt /
ENTRYPOINT ["/go-aqaramqtt"]
