NAME:=gcr.io/distroless/static-debian12
FLAT_NAME=$(shell echo $(NAME) | sed 's/[\/:]/_/g')
FUZZ_TIME:=10

build: chunk diff server fsck

test:
	 go test \
		-timeout 30s \
		-cover \
		github.com/athoune/stream-my-root/pkg/blocks \
		github.com/athoune/stream-my-root/pkg/chunk \
		github.com/athoune/stream-my-root/pkg/trimmed \
		github.com/athoune/stream-my-root/pkg/zero

fuzz-trimmed:
	go test -fuzz=Fuzz -fuzztime $(FUZZ_TIME)s ./pkg/trimmed

fuzz-blocks:
	go test -fuzz=Fuzz -fuzztime $(FUZZ_TIME)s ./pkg/blocks

fuzz: fuzz-trimmed fuzz-blocks

docker:
	docker build -t stream_my_root .

nbd-client:
	docker build -f Dockerfile.client -t nbd-client .

client:
	docker run \
		-ti \
		--rm \
		--cap-add SYS_ADMIN \
		--device /dev/fuse \
		--cap-add SYS_ADMIN \
		--security-opt apparmor:unconfined \
		nbd-client

docker-server:
	docker image build -f Dockerfile.server -t smr-server .

bin:
	mkdir -p bin

chunk: bin
	go build -o bin/chunk cmd/chunk/chunk.go

diff: bin
	go build -o bin/diff cmd/diff/diff.go

fsck: bin
	go build -o bin/fsck cmd/fsck/fsck.go

server: bin
	go build -o bin/server cmd/server/server.go


debug: bin
	go build -o bin/debug cmd/debug/debug.go

img:
	ln -sf $(FLAT_NAME).tar out/the_layer.tar
	docker run \
		-ti \
		--rm \
		-v `pwd`/tar2img.sh:/usr/bin/tar2img.sh \
		-v `pwd`/manifest2layers.sh:/usr/bin/manifest2layers.sh \
		-v `pwd`/out:/work/out \
		-v `pwd`/manifests:/work/manifests \
		-v `pwd`/layers:/work/layers \
		-w /work/ \
		--device /dev/fuse \
		--cap-add SYS_ADMIN \
		--security-opt apparmor:unconfined \
		stream_my_root \
		sh -c 'manifest2layers.sh "$(NAME)" && tar2img.sh "$(NAME)"'

clean:
	rm -rf out
