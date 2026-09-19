package tokenizer

import "sort"

// trieNode is one node of a byte trie. The outgoing edges of a node are the
// contiguous range edges[first:first+count], sorted by byte, which makes the
// children of every node both cache friendly and cheap to search.
type trieNode struct {
	first int32 // index of the first outgoing edge
	count int32 // number of outgoing edges
	value int32 // id of the key ending here, -1 when no key ends here
}

type trieEdge struct {
	b     byte
	child int32
}

// byteTrie is a byte oriented trie. WordPiece splitting uses it for
// longest-prefix lookups and the basic tokenizer uses it to locate the
// special tokens in the input.
type byteTrie struct {
	nodes []trieNode
	edges []trieEdge
	keys  []string // key of every id
}

// linearScanEdges is the widest node child that is searched linearly instead
// of by binary search. Narrow nodes dominate every vocabulary in use: 91% of
// the nodes of the multilingual vocabulary have three or fewer children.
const linearScanEdges = 3

// buildTrie builds a trie holding every key, where keys[i] has id i.
//
// The keys are visited in lexicographic order and split into groups sharing a
// byte, which keeps the children of every node contiguous and sorted. That
// layout is what makes child a binary search over a cache friendly range.
func buildTrie(keys []string) *byteTrie {
	order := make([]int32, len(keys))
	for i := range order {
		order[i] = int32(i)
	}
	sort.Slice(order, func(a, b int) bool {
		ka, kb := keys[order[a]], keys[order[b]]
		if ka != kb {
			return ka < kb
		}
		// Duplicate keys keep the lowest id.
		return order[a] < order[b]
	})
	// Duplicates would break the grouping below, drop them up front.
	unique := order[:0]
	for i, id := range order {
		if i == 0 || keys[order[i-1]] != keys[id] {
			unique = append(unique, id)
		}
	}
	order = unique

	// Pre-size the node and edge slices exactly. The keys are sorted, so a
	// key continues the path of its predecessor up to their common prefix and
	// then adds one node per remaining byte; every node but the root is
	// reached by exactly one edge.
	nodes := 1
	for i, id := range order {
		if i == 0 {
			nodes += len(keys[id])
			continue
		}
		nodes += len(keys[id]) - commonPrefixLen(keys[order[i-1]], keys[id])
	}

	t := &byteTrie{
		nodes: make([]trieNode, 1, nodes),
		edges: make([]trieEdge, 0, nodes-1),
		keys:  keys,
	}
	t.nodes[0].value = -1

	// frame is the key range order[lo:hi], all of whose keys share the trie
	// path to node and are at least depth bytes long.
	type frame struct {
		node  int32
		lo    int32
		hi    int32
		depth int32
	}
	stack := []frame{{node: 0, lo: 0, hi: int32(len(order))}}
	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		lo, hi, depth := int(f.lo), int(f.hi), int(f.depth)
		// A key that stops at this depth is the value of this node. Only the
		// shortest key of a group can do that, and it sorts first.
		if lo < hi && len(keys[order[lo]]) == depth {
			t.nodes[f.node].value = order[lo]
			lo++
		}
		for lo < hi {
			b := keys[order[lo]][depth]
			group := lo
			for group < hi && keys[order[group]][depth] == b {
				group++
			}
			child := int32(len(t.nodes))
			t.nodes = append(t.nodes, trieNode{value: -1})
			// Every edge of a node is appended here, without interleaving,
			// so a node owns the contiguous range that starts at its first
			// edge.
			if t.nodes[f.node].count == 0 {
				t.nodes[f.node].first = int32(len(t.edges))
			}
			t.edges = append(t.edges, trieEdge{b: b, child: child})
			t.nodes[f.node].count++
			stack = append(stack, frame{
				node:  child,
				lo:    int32(lo),
				hi:    int32(group),
				depth: f.depth + 1,
			})
			lo = group
		}
	}
	return t
}

// child returns the node reached by following byte b from node.
func (t *byteTrie) child(node int32, b byte) (int32, bool) {
	n := &t.nodes[node]
	lo, hi := n.first, n.first+n.count
	if n.count <= linearScanEdges {
		for i := lo; i < hi; i++ {
			if e := t.edges[i]; e.b == b {
				return e.child, true
			}
		}
		return 0, false
	}
	for lo < hi {
		mid := int(lo+hi) / 2
		switch e := t.edges[mid]; {
		case e.b == b:
			return e.child, true
		case e.b < b:
			lo = int32(mid) + 1
		default:
			hi = int32(mid)
		}
	}
	return 0, false
}

// exact returns the id of key.
func (t *byteTrie) exact(key string) (int32, bool) {
	node := int32(0)
	for i := 0; i < len(key); i++ {
		child, ok := t.child(node, key[i])
		if !ok {
			return 0, false
		}
		node = child
	}
	if t.nodes[node].value < 0 {
		return 0, false
	}
	return t.nodes[node].value, true
}

// longestPrefix returns the id of the longest key that is a prefix of s.
func (t *byteTrie) longestPrefix(s string) (int32, bool) {
	best := int32(-1)
	node := int32(0)
	for i := 0; i < len(s); i++ {
		child, ok := t.child(node, s[i])
		if !ok {
			break
		}
		node = child
		if t.nodes[node].value >= 0 {
			best = t.nodes[node].value
		}
	}
	return best, best >= 0
}

// longestPrefixParts returns the id of the longest key that is a prefix of
// prefix+suffix and starts with prefix. WordPiece uses it for continuation
// pieces, whose "##" marker would otherwise have to be concatenated into a new
// string for every attempt.
func (t *byteTrie) longestPrefixParts(prefix, suffix string) (int32, bool) {
	best := int32(-1)
	node := int32(0)
	for i := 0; i < len(prefix); i++ {
		child, ok := t.child(node, prefix[i])
		if !ok {
			return -1, false
		}
		node = child
		if t.nodes[node].value >= 0 {
			best = t.nodes[node].value
		}
	}
	for i := 0; i < len(suffix); i++ {
		child, ok := t.child(node, suffix[i])
		if !ok {
			break
		}
		node = child
		if t.nodes[node].value >= 0 {
			best = t.nodes[node].value
		}
	}
	return best, best >= 0
}

// prefixes calls fn for every key that is a prefix of s, shortest first, and
// stops after at most limit keys. It stops early when fn returns false.
func (t *byteTrie) prefixes(s string, limit int, fn func(id int32) bool) {
	if limit <= 0 {
		return
	}
	node := int32(0)
	n := 0
	if t.nodes[0].value >= 0 {
		if !fn(t.nodes[0].value) {
			return
		}
		n++
	}
	for i := 0; i < len(s) && n < limit; i++ {
		child, ok := t.child(node, s[i])
		if !ok {
			return
		}
		node = child
		if t.nodes[node].value >= 0 {
			if !fn(t.nodes[node].value) {
				return
			}
			n++
		}
	}
}

func commonPrefixLen(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return i
}
