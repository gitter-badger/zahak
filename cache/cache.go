package cache

type CachedEval struct {
	Hash  uint64
	Eval  int32
	Depth int8
	Type  NodeType
	Age   uint16
}

type NodeType uint8

const (
	Exact      NodeType = 1 << iota // PV-Node
	UpperBound                      // All-Node
	LowerBound                      // Cut-Node
	Missing
)

var oldAge = uint16(5)

const CACHE_ENTRY_SIZE = uint32(64 + 16 + 8 + 8 + 32)

type Cache struct {
	hashes   []uint64
	evals    []int32
	depths   []int8
	types    []NodeType
	ages     []uint16
	size     uint32
	consumed int
}

func (c *Cache) Consumed() int {
	return int((float64(c.consumed) / float64(len(c.hashes))) * 1000)
}

var EmptyCache = Cache{nil, nil, nil, nil, nil, 0, 0}
var TranspositionTable Cache = EmptyCache

func (c *Cache) hash(key uint64) uint32 {
	return uint32(key>>32) % c.size
}

func (c *Cache) insert(key uint32, hash uint64, eval int32, depth int8, tpe NodeType, age uint16) {
	c.hashes[key] = hash
	c.evals[key] = eval
	c.depths[key] = depth
	c.types[key] = tpe
	c.ages[key] = age
}

func (c *Cache) Set(hash uint64, eval int32, depth int8, tpe NodeType, age uint16) {
	key := c.hash(hash)
	if c.types[key] != Missing {
		if hash == c.hashes[key] {
			c.insert(key, hash, eval, depth, tpe, age)
			return
		}
		if age-c.ages[key] >= oldAge {
			c.insert(key, hash, eval, depth, tpe, age)
			return
		}
		if c.depths[key] > depth {
			return
		}
		if c.types[key] == Exact || tpe != Exact {
			return
		} else if tpe == Exact {
			c.insert(key, hash, eval, depth, tpe, age)
			return
		}
		c.insert(key, hash, eval, depth, tpe, age)
	} else {
		c.consumed += 1
		c.insert(key, hash, eval, depth, tpe, age)
	}
}

func (c *Cache) Get(hash uint64) (int32, int8, NodeType, bool) {
	key := c.hash(hash)
	storedHash := c.hashes[key]
	if storedHash == hash {
		return c.evals[key], c.depths[key], c.types[key], true
	}
	return -1, -1, Missing, false
}

func NewCache(megabytes uint32) {
	size := megabytes * 1024 * 1024 / CACHE_ENTRY_SIZE
	hashes := make([]uint64, size)
	evals := make([]int32, size)
	depths := make([]int8, size)
	types := make([]NodeType, size)
	ages := make([]uint16, size)
	TranspositionTable = Cache{hashes, evals, depths, types, ages, uint32(size), 0} //s, current: 0}
	for i := 0; i < int(size); i++ {
		TranspositionTable.types[i] = Missing
	}
}

func ResetCache() {
	if TranspositionTable.size != EmptyCache.size {
		size := TranspositionTable.size
		TranspositionTable.hashes = make([]uint64, size)
		TranspositionTable.evals = make([]int32, size)
		TranspositionTable.depths = make([]int8, size)
		TranspositionTable.types = make([]NodeType, size)
		TranspositionTable.ages = make([]uint16, size)
		TranspositionTable.consumed = 0
		for i := 0; i < int(TranspositionTable.size); i++ {
			TranspositionTable.types[i] = Missing
		}
	} else {
		NewCache(400)
	}
}
