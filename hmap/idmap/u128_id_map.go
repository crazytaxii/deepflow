package idmap

import (
	"sync"
)

type u128IDMapNode struct {
	key0  uint64
	key1  uint64
	value uint32

	next int32 // 表示节点所在冲突链的下一个节点的 buffer 数组下标
	slot int32 // 记录 node 对应的哈希 slot ，为了避免 Clear 函数遍历整个 slotHead 数组
}

func (n *u128IDMapNode) equal(key0, key1 uint64) bool {
	return n.key0 == key0 && n.key1 == key1
}

var blankU128MapNodeForInit u128IDMapNode

type u128IDMapNodeBlock []u128IDMapNode

var u128IDMapNodeBlockPool = sync.Pool{New: func() interface{} {
	return u128IDMapNodeBlock(make([]u128IDMapNode, _BLOCK_SIZE))
}}

// 注意：不是线程安全的
type U128IDMap struct {
	buffer []u128IDMapNodeBlock // 存储Map节点，以矩阵的方式组织，提升内存申请释放效率

	slotHead []int32 // 哈希桶，slotHead[i] 表示哈希值为 i 的冲突链的第一个节点为 buffer[[ slotHead[i]] ]]
	size     int     // buffer中存储的有效节点总数
	width    int     // 哈希桶中最大冲突链长度
}

func NewU128IDMap(hashSlots uint32) *U128IDMap {
	if hashSlots >= 1<<30 {
		panic("hashSlots is too large")
	}

	i := uint32(1)
	for i < hashSlots {
		i <<= 1
	}
	hashSlots = i

	m := &U128IDMap{
		buffer:   make([]u128IDMapNodeBlock, 0),
		slotHead: make([]int32, hashSlots),
	}

	for i := uint32(0); i < hashSlots; i++ {
		m.slotHead[i] = -1
	}
	return m
}

func (m *U128IDMap) Size() int {
	return m.size
}

func (m *U128IDMap) Width() int {
	return m.width
}

// 第一个返回值表示value，第二个返回值表示是否进行了Add。若key已存在，指定overwrite=true可覆写value。
func (m *U128IDMap) AddOrGet(key0, key1 uint64, value uint32, overwrite bool) (uint32, bool) {
	slot := (uint32(key0>>32) ^ uint32(key0) ^ uint32(key1>>32) ^ uint32(key1)) & uint32(len(m.slotHead)-1)
	head := m.slotHead[slot]

	width := 0
	next := head
	for next != -1 {
		width++
		node := &m.buffer[next>>_BLOCK_SIZE_BITS][next&_BLOCK_SIZE_MASK]
		if node.equal(key0, key1) {
			if overwrite {
				node.value = value
			} else {
				value = node.value
			}
			return value, false
		}
		next = node.next
	}

	if m.size >= len(m.buffer)<<_BLOCK_SIZE_BITS { // expand
		m.buffer = append(m.buffer, u128IDMapNodeBlockPool.Get().(u128IDMapNodeBlock))
	}
	node := &m.buffer[m.size>>_BLOCK_SIZE_BITS][m.size&_BLOCK_SIZE_MASK]
	node.key0 = key0
	node.key1 = key1
	node.value = value
	node.next = head
	node.slot = int32(slot)

	m.slotHead[slot] = int32(m.size)
	m.size++

	if m.width < width+1 {
		m.width = width + 1
	}

	return value, true
}

func (m *U128IDMap) Get(key0, key1 uint64) (uint32, bool) {
	slot := (uint32(key0>>32) ^ uint32(key0) ^ uint32(key1>>32) ^ uint32(key1)) & uint32(len(m.slotHead)-1)
	head := m.slotHead[slot]

	next := head
	for next != -1 {
		node := &m.buffer[next>>_BLOCK_SIZE_BITS][next&_BLOCK_SIZE_MASK]
		if node.equal(key0, key1) {
			return node.value, true
		}
		next = node.next
	}
	return 0, false
}

func (m *U128IDMap) Clear() {
	for i := 0; i < m.size; i += _BLOCK_SIZE {
		for j := 0; j < _BLOCK_SIZE && i+j < m.size; j++ {
			node := &m.buffer[i>>_BLOCK_SIZE_BITS][j]
			m.slotHead[node.slot] = -1
			*node = blankU128MapNodeForInit
		}
		u128IDMapNodeBlockPool.Put(m.buffer[i>>_BLOCK_SIZE_BITS])
		m.buffer[i>>_BLOCK_SIZE_BITS] = nil
	}

	m.buffer = m.buffer[:0]

	m.size = 0
	m.width = 0
}
