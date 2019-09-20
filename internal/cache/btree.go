package cache

type btree struct {
	Root *node
}

type node struct {
	Left   *node
	Right  *node
	Key    int64
	Values map[string]struct{}
}

func (b *btree) insert(key int64, value string) {
	n := &node{
		Left:   nil,
		Right:  nil,
		Key:    key,
		Values: map[string]struct{}{value: {}},
	}

	if b.Root == nil {
		b.Root = n
		return
	}

	b.Root.insert(n)
}

func (b *btree) removeNodesLessThan(k int64) map[string]struct{} {
	if b.Root == nil {
		return nil
	}

	fakeParent := &node{Right: b.Root}
	rm := b.Root.removeNodesOlderThan(k, fakeParent)
	b.Root = fakeParent.Right

	var nodes []*node
	b.traverse(b.Root, func(node *node) {
		nodes = append(nodes, node)
	})

	b.Root = nil
	b.rebalance(nodes, 0, len(nodes)-1)

	return rm
}

func (b *btree) traverse(n *node, f func(*node)) {
	if n == nil {
		return
	}
	b.traverse(n.Left, f)
	f(n)
	b.traverse(n.Right, f)
}

func (b *btree) rebalance(nodes []*node, start, end int) {
	m := (start + end) / 2
	if m >= len(nodes) || start > end {
		return
	}

	if b.Root == nil {
		b.Root = nodes[m]
		b.Root.Values = nodes[m].Values
	} else {
		b.Root.insert(nodes[m])
	}

	b.rebalance(nodes, start, m-1)
	b.rebalance(nodes, m+1, end)
}

func (n *node) insert(newNode *node) {
	if n.Key > newNode.Key {
		if n.Left == nil {
			n.Left = newNode
			return
		}
		n.Left.insert(newNode)
		return
	}

	if n.Key < newNode.Key {
		if n.Right == nil {
			n.Right = newNode
			return
		}
		n.Right.insert(newNode)
		return
	}

	if n.Key == newNode.Key {
		for k, v := range newNode.Values {
			n.Values[k] = v
		}
	}
}

func (n *node) removeNodesOlderThan(k int64, parent *node) map[string]struct{} {
	if n == nil {
		return nil
	}

	if k < n.Key {
		return n.Left.removeNodesOlderThan(k, n)
	}

	keys := make(map[string]struct{})
	for k := range n.Values {
		keys[k] = struct{}{}
	}

	if n.Left != nil {
		for k := range n.Left.removeNodesOlderThan(k, n) {
			keys[k] = struct{}{}
		}
	}

	if n.Right != nil {
		for k := range n.Right.removeNodesOlderThan(k, n) {
			keys[k] = struct{}{}
		}
	}

	if parent.Right == n {
		parent.Right = n.Right
	}

	if parent.Left == n {
		parent.Left = n.Left
	}

	return keys
}
