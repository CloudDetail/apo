package model

import (
	"encoding/json"
	"fmt"
	"github.com/magiconair/properties/assert"
	"testing"
)

type Flame struct {
	Version     int         `json:"version"`
	Flamebearer FlameBearer `json:"flamebearer"`
	Metadata    Metadata    `json:"metadata"`
}

type Metadata struct {
	Name       string `json:"name"`
	Format     string `json:"format"`
	SampleRate int    `json:"sampleRate"`
	SpyName    string `json:"spyName"`
	Units      string `json:"units"`
}

func TestMergeFlameGraph(t *testing.T) {
	root := newNode("root", 0, 60)
	func1 := newNode("func1", 20, 60)
	func2 := newNode("func2", 30, 30)
	func3 := newNode("func3", 10, 10)
	root.children = append(root.children, func1)
	func1.children = append(func1.children, func2, func3)
	tree := &Tree{root: []*node{root}}

	flame := NewFlameGraph(tree, 0)
	oldFlame := NewFlameGraph(tree, 0)
	mergeTree := &Tree{}
	mergeTree.MergeFlameGraph(flame)
	output := NewFlameGraph(mergeTree, 0)
	assert.Equal(t, output, oldFlame)
}

// TestLength 测试下层节点的长度大于上层
// 结果：&{[total root func1 func2] [[0 50 0 0] [0 50 0 1] [0 50 20 2] [20 60 60 3]] 50 60}
// 下层大于上层，展示时显示为120%，点击之后下层比上层宽
func TestLength(t *testing.T) {
	root := newNode("root", 0, 50)
	func1 := newNode("func1", 20, 50)
	// 下层节点比上层大，属于错误数据
	func2 := newNode("func2", 60, 60)
	root.children = append(root.children, func1)
	func1.children = append(func1.children, func2)
	tree := &Tree{root: []*node{root}}
	flame := NewFlameGraph(tree, 0)
	fmt.Println(flame)
	fmt.Println(wrapToFlame(*flame))
}

// TestFilterNodeNum 测试maxNodes过滤节点数量
func TestFilterNodeNum(t *testing.T) {
	nodeNum := 10
	nodes := make([]*node, 10)
	root := newNode("root", 0, 0)
	for i := 0; i < nodeNum; i++ {
		var size = int64(i*10 + 10)
		nodes[i] = newNode(fmt.Sprintf("func%d", i), size, size)
		root.children = append(root.children, nodes[i])
		root.total += size
	}
	tree := &Tree{root: []*node{root}}
	flame := NewFlameGraph(tree, 5)
	expect := &FlameBearer{
		Names: []string{"total", "root", "other", "func9", "func8", "func7", "func6"},
		Levels: [][]int64{
			{0, 550, 0, 0}, {0, 550, 0, 1}, {0, 70, 70, 6, 0, 80, 80, 5, 0, 90, 90, 4, 0, 100, 100, 3, 0, 210, 210, 2},
		},
		NumTicks: 550,
		MaxSelf:  210,
	}
	assert.Equal(t, flame, expect)
	fmt.Println(wrapToFlame(*flame))
}

// TestFlameGroup
// 1. 测试两个节点的子节点不同（1->2 3->4 | 1->4 2->3）√
// 2. 测试偏移相同的节点是否会被覆盖 (1->2 offset30 | 1->3 offset30) √
// 3. 测试相同节点合并（1->2 | 1->2） √
// 4. 测试同级节点顺序不同（1->2,3 | 1->3,2）√
// 5. 测试层级数量不同（1->2 | 1->2->3）√
// 6. 测试两个完全不同的火焰图（1->2 | 3->4）√
func TestFlameGroup(t *testing.T) {
	root := newNode("root", 0, 60)
	func1 := newNode("func1", 20, 60)
	func2 := newNode("func2", 30, 30)
	func3 := newNode("func3", 10, 10)
	root.children = append(root.children, func1)
	func1.children = append(func1.children, func2, func3)
	tree := &Tree{root: []*node{root}}
	flame1 := NewFlameGraph(tree, 0)
	// 交换，func3 offset与flame1 func2 offset相同
	func1.children[0], func1.children[1] = func3, func2
	flame2 := NewFlameGraph(tree, 0)
	mergeTree := &Tree{}
	mergeTree.MergeFlameGraph(flame1)
	mergeTree.MergeFlameGraph(flame2)
	got := NewFlameGraph(mergeTree, 0)
	/*
		tree1:
			root 0 0 60
			func1 0 20 60
			func2 20 30 30 func3 0 10 10
		tree2:
			root 0 0 60
			func1 0 20 60
			func3 20 10 10 func2 0 30 30
		expect:
			root 0 0 120 0
			func1 0 40 120 1
			func2 40 60 60 2 func3 0 20 20 3
	*/
	expect := &FlameBearer{
		Names: []string{"total", "root", "func1", "func3", "func2"},
		Levels: [][]int64{
			{0, 120, 0, 0}, {0, 120, 0, 1}, {0, 120, 40, 2}, {40, 60, 60, 4, 0, 20, 20, 3},
		},
		NumTicks: 120,
		MaxSelf:  60,
	}
	assert.Equal(t, got, expect)
	fmt.Println(wrapToFlame(*got))
}

func wrapToFlame(flame FlameBearer) string {
	f := Flame{
		Version:     1,
		Flamebearer: flame,
		Metadata:    Metadata{Format: "single"},
	}
	fStr, _ := json.Marshal(f)
	return string(fStr)
}

func newNode(name string, self, total int64) *node {
	return &node{
		name:  name,
		self:  self,
		total: total,
	}
}

func TestToPyroscopeFlameBearer(t *testing.T) {
	data := "{\"names\":[\"total\",\"other\",\"libpthread-2.19.so.start_thread\",\"./usr/lib/jvm/java-8-openjdk-amd64/jre/lib/amd64/server/libjvm.so\",\"libpthread-2.19.so.pthread_cond_wait@@GLIBC_2.3.2\",\"java/lang/Thread.run\",\"org/apache/tomcat/util/threads/TaskThread$WrappingRunnable.run\",\"java/util/concurrent/ThreadPoolExecutor$Worker.run\",\"java/util/concurrent/ThreadPoolExecutor.runWorker\",\"java/util/concurrent/ThreadPoolExecutor.getTask\",\"org/apache/tomcat/util/threads/TaskQueue.take\",\"java/util/concurrent/LinkedBlockingQueue.take\",\"java/util/concurrent/locks/AbstractQueuedSynchronizer$ConditionObject.await\",\"java/util/concurrent/locks/LockSupport.park\",\"sun/misc/Unsafe.park\"],\"levels\":[[0,4699450000000,0,0],[0,2269450000000,0,5,0,1709840000000,0,2,0,720160000000,720160000000,1],[0,1079270000000,0,6,0,1190180000000,1190180000000,1,0,1619900000000,0,3,0,89940000000,89940000000,1],[0,1079270000000,0,7,1190180000000,1619900000000,0,3],[0,1079270000000,10000000,8,1190180000000,1619900000000,30000000,3],[10000000,1074660000000,0,9,0,4600000000,4600000000,1,1190210000000,1439810000000,30000000,3,0,180060000000,180060000000,1],[10000000,894760000000,0,10,0,179900000000,179900000000,1,1194840000000,1349780000000,0,3,0,90000000000,90000000000,1],[10000000,894760000000,0,10,1374740000000,809820000000,809820000000,4,0,539960000000,539960000000,1],[10000000,894760000000,0,11],[10000000,894760000000,0,12],[10000000,894760000000,0,13],[10000000,894760000000,0,14],[10000000,894760000000,0,3],[10000000,894760000000,894760000000,4]],\"numTicks\":4699450000000,\"maxSelf\":1190180000000}"
	var raw FlameBearer
	err := json.Unmarshal([]byte(data), &raw)
	if err != nil {
		t.Fatal(err)
	}
	jsonStr := wrapToFlame(raw)
	fmt.Println(jsonStr)
}
