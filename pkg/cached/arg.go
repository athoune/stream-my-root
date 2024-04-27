package cached

import "encoding/binary"

type SetArg struct {
	Key  string
	Size uint32
}

func (s *SetArg) UnmarshalBinary(data []byte) error {
	s.Size = binary.BigEndian.Uint32(data[0:4])
	s.Key = string(data[4:])
	return nil
}

func (s *SetArg) MarshalBinary() (data []byte, err error) {
	bkey := []byte(s.Key)
	arg := make([]byte, len(bkey)+4)
	binary.BigEndian.PutUint32(arg[0:4], s.Size)
	copy(arg[4:], bkey)
	return arg, nil
}

type Bool bool

func (b *Bool) UnmarshalBinary(data []byte) error {
	if data[0] == 0 {
		*b = false
	} else {
		*b = true
	}
	return nil
}

func (b Bool) MarshalBinary() (data []byte, err error) {
	if b {
		return []byte{1}, nil
	}
	return []byte{0}, nil
}
