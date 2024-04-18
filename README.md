# Stream my root

[On-demand Container Loading in AWS Lambda](https://arxiv.org/abs/2305.13162)


## Test it

Build the make_ext4fs image

```bash
cd contrib/resurrected-make-ext4fs
make
```

Build the tool image

```bash
make docker
```

Fetch some images

```bash
make img NAME=gcr.io/distroless/python3-debian12
make img NAME=gcr.io/distroless/base-debian12
# *.img files are stored in out/
ls out
```

Build tools (with golang)

```bash
make
```

Chunk images

```bash
./bin/chunk out/*.img
# recipe are stored near the img file
ls out
# chunks are stored is smr/
ls smr
```

Run the server

```bash
# the first arg is a recipe
./bin/server out/gcr.io_distroless_python3-debian12.img.recipe
```

Mount the image (from a Linux)

```bash
# split your tmux with ctrl-% and watch the kernel yelling at nbd
tail -f /var/log/kern.log
sudo nbd-client -N smr localhost 10809 /dev/nbd1
sudo mkdir /mnt/smr
sudo mount -o ro -t ext4 /dev/nbd1 /mnt/smr
ls /mnt/smr
```

## Stuff to read

* https://github.com/opencontainers/image-spec/blob/main/layer.md
* https://reproducible-builds.org/docs/system-images/
