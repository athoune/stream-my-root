NAME:=gcr.io/distroless/static-debian12

build:
	mkdir -p bin
	go build -o bin/chunk chunk.go

venv:
	python3 -m venv venv
	./venv/bin/pip install -U pip wheel

venv/bin/docker-squash: venv
	./venv/bin/pip install docker-squash

squash: venv/bin/docker-squash

out/img.tar: squash
	mkdir -p out
	./venv/bin/docker-squash $(NAME)  --output-path out/img.tar

out/the_layer.tar: out/img.tar
	cd out && \
	tar -xvf img.tar && \
	ln -sf `find . -name layer.tar | head -n1` the_layer.tar

img: out/the_layer.tar
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
# make_ext4fs -l 1G -b 64k -L stream -g 256 toto.img

clean:
	rm -rf out
