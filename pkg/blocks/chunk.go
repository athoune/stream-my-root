package blocks

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
)

var encoder, _ = zstd.NewWriter(nil)

type BlockVistor func(start int64, content []byte, sha [32]byte) error

func VisitBlock(f io.Reader, block_size int, visitor BlockVistor) (int, error) {
	buffer := make([]byte, block_size)
	poz := 0
	var block []byte
	for {
		_, err := f.Read(buffer)
		poz += 1
		if err != nil {
			if err == io.EOF { // It's not really an error, but the file is completly read
				return poz * block_size, nil
			}
			return poz * block_size, err
		}
		block = Rtrim(buffer)
		if len(block) > 0 {
			err = visitor(int64(poz-1), block, sha256.Sum256(buffer))
			if err != nil {
				return poz * block_size, err
			}
		}
	}
}

func ChunkRawFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("can't open %s : %s", path, err)
	}
	defer f.Close()
	r, err := os.OpenFile(fmt.Sprintf("%s.recipe", path), os.O_RDWR+os.O_CREATE, 0640)
	if err != nil {
		return fmt.Errorf("can't open recipe : %s", err)
	}
	defer r.Close()

	v := func(start int64, content []byte, sha [32]byte) error {
		h := hex.EncodeToString(sha[:])
		_, err = fmt.Fprintf(r, "%d %s\n", start, h)
		if err != nil {
			return err
		}
		p := fmt.Sprintf("smr/%s.zst", h)
		_, err = os.Stat(p)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			f, err = os.OpenFile(p, os.O_CREATE+os.O_WRONLY, 0640)
			if err != nil {
				return err
			}
			_, err = f.Write(encoder.EncodeAll(content, make([]byte, 0, len(content))))
			if err != nil {
				return err
			}
		}
		return nil
	}
	_, err = VisitBlock(f, 512*1024, v)
	if err != nil {
		return fmt.Errorf("visiting block : %s", err)
	}
	return nil
}
