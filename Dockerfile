FROM make_ext4
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        jq \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
ARG CRANE_VERSION=v0.19.1
ENV OS=Linux
ENV ARCH=x86_64
RUN mkdir -p /usr/local/bin \
    && curl -sL -o go-containerregistry.tar.gz "https://github.com/google/go-containerregistry/releases/download/${CRANE_VERSION}/go-containerregistry_${OS}_${ARCH}.tar.gz" \
    && tar -zxvf go-containerregistry.tar.gz -C /usr/local/bin/ crane
