NAME:=gcr.io/distroless/static-debian12
FLAT_NAME=$(shell echo $(NAME) | sed 's/[\/:]/_/g')

build: chunk diff server fsck

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
