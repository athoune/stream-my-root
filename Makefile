NAME:=gcr.io/distroless/static-debian12
FLAT_NAME=$(shell echo $(NAME) | sed 's/[\/:]/_/g')

build: chunk diff

bin:
	mkdir -p bin

chunk: bin
	go build -o bin/chunk cmd/chunk/chunk.go

diff: bin
	go build -o bin/diff cmd/diff/diff.go

venv:
	python3 -m venv venv
	./venv/bin/pip install -U pip wheel

venv/bin/docker-squash: venv
	./venv/bin/pip install docker-squash

squash: venv/bin/docker-squash

out/$(FLAT_NAME).tar: squash
	mkdir -p out/squashed
	./venv/bin/docker-squash $(NAME)  --output-path out/img.tar
	cd out/squashed && \
	tar -xvf ../img.tar
	mv `find out/squashed -name layer.tar | head -n1` out/$(FLAT_NAME).tar
	rm -rf out/squashed
	rm out/img.tar

img: out/$(FLAT_NAME).tar
	ln -sf $(FLAT_NAME).tar out/the_layer.tar
	docker run \
		-ti \
		--rm \
		-v `pwd`/tar2img.sh:/usr/bin/tar2img.sh \
		-v `pwd`/out:/out \
		-w /out \
		--device /dev/fuse \
		--cap-add SYS_ADMIN \
		--security-opt apparmor:unconfined \
		make_ext4 \
		tar2img.sh
	rm out/the_layer.tar
	mv out/root.img out/$(FLAT_NAME).img

clean:
	rm -rf out
