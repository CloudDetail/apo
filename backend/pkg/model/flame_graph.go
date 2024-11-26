// copied from https://github.com/grafana/pyroscope
package model

import (
	"sort"
)

type FlameBearer struct {
	Names    []string  `json:"names"`
	Levels   [][]int64 `json:"levels"`
	NumTicks int64     `json:"numTicks"`
	MaxSelf  int64     `json:"maxSelf"`
}

type Tree struct {
	root []*node
}

type node struct {
	parent      *node
	children    []*node
	self, total int64
	name        string
}

type Stack[T any] struct {
	values []T
}

func (s *Stack[T]) Push(v T) {
	s.values = append(s.values, v)
}

func (s *Stack[T]) Pop() (res T, ok bool) {
	if len(s.values) == 0 {
		ok = false
		return
	}
	top := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return top, true
}

func (s *Stack[T]) Slice() []T {
	result := make([]T, 0, len(s.values))
	for i := len(s.values) - 1; i >= 0; i-- {
		result = append(result, s.values[i])
	}
	return result
}

type stackNode struct {
	xOffset int
	level   int
	node    *node
}

func (t *Tree) MergeFlameGraph(src *FlameBearer) {
	for _, l := range src.Levels {
		prev := int64(0)
		for i := 0; i < len(l); i += 4 {
			delta := l[i] + l[i+1]
			l[i] += prev
			prev += delta
		}
	}
	dst := make([]string, 0, len(src.Levels))
	for i, l := range src.Levels {
		if i == 0 {
			continue
		}
		for j := 0; j < len(l); j += 4 {
			self := l[j+2]
			if self > 0 {
				dst = buildStack(dst, src, i, j)
				t.InsertStack(self, dst...)
			}
		}
	}
}

func buildStack(dst []string, f *FlameBearer, level, idx int) []string {
	if cap(dst) < level {
		dst = make([]string, level, level*2)
	} else {
		dst = dst[:level]
	}
	dst[level-1] = f.Names[f.Levels[level][idx+3]]
	x := f.Levels[level][idx]
	for i := level - 1; i > 0; i-- {
		j := sort.Search(len(f.Levels[i])/4, func(j int) bool { return f.Levels[i][j*4] > x }) - 1
		dst[i-1] = f.Names[f.Levels[i][j*4+3]]
		x = f.Levels[i][j*4]
	}
	return dst
}

func (t *Tree) InsertStack(v int64, stack ...string) {
	if v <= 0 {
		return
	}
	r := &node{children: t.root}
	n := r
	for s := range stack {
		name := stack[s]
		n.total += v
		// Inlined node.insert
		i, j := 0, len(n.children)
		for i < j {
			h := int(uint(i+j) >> 1)
			if n.children[h].name < name {
				i = h + 1
			} else {
				j = h
			}
		}
		if i < len(n.children) && n.children[i].name == name {
			n = n.children[i]
		} else {
			child := &node{parent: n, name: name}
			n.children = append(n.children, child)
			copy(n.children[i+1:], n.children[i:])
			n.children[i] = child
			n = child
		}
	}
	// Leaf.
	n.total += v
	n.self += v
	t.root = r.children
}

func NewFlameGraph(t *Tree) *FlameBearer {
	var total, max int64
	for _, node := range t.root {
		total += node.total
	}
	names := []string{}
	nameLocationCache := map[string]int{}
	res := []*Stack[int64]{}

	stack := Stack[stackNode]{}
	stack.Push(stackNode{xOffset: 0, level: 0, node: &node{children: t.root, total: total}})
	for {
		current, hasMoreNodes := stack.Pop()
		if !hasMoreNodes {
			break
		}
		if current.node.self > max {
			max = current.node.self
		}
		var i int
		var ok bool
		name := current.node.name
		if i, ok = nameLocationCache[name]; !ok {
			i = len(names)
			if i == 0 {
				name = "total"
			}
			nameLocationCache[name] = i
			names = append(names, name)
		}

		if current.level == len(res) {
			s := &Stack[int64]{}
			res = append(res, s)
		}

		// i+0 = x offset
		// i+1 = total
		// i+2 = self
		// i+3 = index in names array
		level := res[current.level]
		level.Push(int64(i))
		level.Push(current.node.self)
		level.Push(current.node.total)
		level.Push(int64(current.xOffset))
		current.xOffset += int(current.node.self)

		otherTotal := int64(0)
		for _, child := range current.node.children {
			if child.name != "other" {
				stack.Push(stackNode{xOffset: current.xOffset, level: current.level + 1, node: child})
				current.xOffset += int(child.total)
			} else {
				otherTotal += child.total
			}
		}
		if otherTotal != 0 {
			child := &node{
				name:   "other",
				parent: current.node,
				self:   otherTotal,
				total:  otherTotal,
			}
			stack.Push(stackNode{xOffset: current.xOffset, level: current.level + 1, node: child})
			current.xOffset += int(child.total)
		}
	}

	result := make([][]int64, len(res))
	for i := range result {
		result[i] = res[i].Slice()
	}
	// delta encode xoffsets
	for _, l := range result {
		prev := int64(0)
		for i := 0; i < len(l); i += 4 {
			l[i] -= prev
			prev += l[i] + l[i+1]
		}
	}
	levels := make([][]int64, len(result))
	for i := range levels {
		levels[i] = result[i]
	}

	return &FlameBearer{
		Names:    names,
		Levels:   levels,
		NumTicks: total,
		MaxSelf:  max,
	}
}
