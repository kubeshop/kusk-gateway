FROM envoyproxy/envoy-dev:latest
# Disable dpkg prompts
ENV DEBIAN_FRONTEND=noninteractive
RUN apt update -y -qq&& \
    apt install -y curl && \
    curl --silent -Lk -o /usr/local/bin/gomplate https://github.com/hairyhenderson/gomplate/releases/download/v3.10.0/gomplate_linux-amd64 &&\
    chmod 755 /usr/local/bin/gomplate && \
    apt clean -y &&\
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY docker-entrypoint.sh /
ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD ["envoy", "-c" ,"/etc/envoy/envoy.yaml"]