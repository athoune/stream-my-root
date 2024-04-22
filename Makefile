NAME:=gcr.io/distroless/static-debian12
FLAT_NAME=$(shell echo $(NAME) | sed 's/[\/:]/_/g')
FUZZ_TIME:=10
ifeq "$(shell uname)" "Darwin"
ARCH_UNAME:=$(shell uname -a | cut -f 15 -d ' ')
else
ARCH_UNAME:=$(shell uname -a | cut -f 12 -d ' ')
endif
ifeq "$(ARCH_UNAME)" "x86_64"
ARCH:=amd64
endif
ifeq "$(ARCH_UNAME)" "aarch64"
ARCH:=arm64
endif
CRANE_ARCH:=$(ARCH)
ifeq "$(CRANE_ARCH)" "amd64"
CRANE_ARCH=x86_64
endif

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

make_ext4fs: contrib/resurrected-make-ext4fs/Makefile
	cd contrib/resurrected-make-ext4fs && make

contrib/resurrected-make-ext4fs/Makefile:
	make submodule

submodule:
	git submodule init
	git submodule update

docker-tool: make_ext4fs
	docker build \
		-t stream_my_root \
		--build-arg "ARCH=$(ARCH)" \
		--build-arg "ARCH_UNAME=$(ARCH_UNAME)" \
		--build-arg "CRANE_ARCH=$(CRANE_ARCH)" \
		.

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

docker-client:
	docker build -f Dockerfile.client -t nbd-client .

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

lima-create:
	limactl create --cpus=2 --memory=2 --name=smr-debian template://debian

lima:
	limactl start smr-debian
	limactl shell smr-debian

clean:
	rm -rf out
	rm -rf smr
	rm -rf layers
