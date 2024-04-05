FROM make_ext4
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*\
    && mkdir -p /usr/local/bin
ENV OS=Linux
ENV ARCH=x86_64
ENV JQ_VERSION=1.7.1
# FIXME assert sha256 for downloaded files
# FIXME Debian jq is 1.6 finding how to handle list with this version must be better
RUN curl -o /usr/local/bin/jq -sL https://github.com/jqlang/jq/releases/download/jq-${JQ_VERSION}/jq-linux-amd64 \
    && chmod +x /usr/local/bin/jq
ARG CRANE_VERSION=v0.19.1
RUN curl -sL -o go-containerregistry.tar.gz "https://github.com/google/go-containerregistry/releases/download/${CRANE_VERSION}/go-containerregistry_${OS}_${ARCH}.tar.gz" \
    && tar -zxvf go-containerregistry.tar.gz -C /usr/local/bin/ crane
