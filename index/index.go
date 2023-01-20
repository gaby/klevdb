package index

import (
	"hash/fnv"

	"github.com/klev-dev/klevdb/message"
)

type Item struct {
	Offset    int64
	Position  int64
	Timestamp int64
	KeyHash   uint64
}

type Params struct {
	Times bool
	Keys  bool
}

func (o Params) keyOffset() int {
	off := 8 + 8 // offset + position
	if o.Times {
		off += 8
	}
	return off
}

func (o Params) Size() int64 {
	sz := int64(8 + 8) // offset + position
	if o.Times {
		sz += 8
	}
	if o.Keys {
		sz += 8
	}
	return sz
}

func (o Params) NewItem(m message.Message, position int64) Item {
	it := Item{Offset: m.Offset, Position: position}

	if o.Times {
		it.Timestamp = m.Time.UnixMicro()
	}

	if o.Keys {
		hasher := fnv.New64a()
		hasher.Write(m.Key)
		it.KeyHash = hasher.Sum64()
	}

	return it
}