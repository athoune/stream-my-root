package blocks

type Slice struct {
	blocks     []*Block
	start, end int64
}

func (s *Slice) Hashes() []string {
	hashes := make([]string, len(s.blocks))
	var h string
	for i, block := range s.blocks {
		if block == nil {
			h = ""
		} else {
			h = block.Hash
		}
		hashes[i] = h
	}
	return hashes
}

func (s *Slice) Length() int {
	return len(s.blocks)
}
