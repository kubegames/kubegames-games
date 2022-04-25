package gamelogic

import (
	"fmt"
	"strconv"
)

//定义二叉树的节点
type Node struct {
	value int
	hight int
	left  *Node
	mid   *Node
	right *Node
}

//功能：打印节点的值
//参数：nil
//返回值：nil
func (node *Node) Print() {
	fmt.Printf("%d ", node.value)
}

//功能：设置节点的值
//参数：节点的值
//返回值：nil
func (node *Node) SetValue(value int) {
	node.value = value
}

//功能：创建节点
//参数：节点的值
//返回值：nil
func CreateNode(value int, hight int) *Node {
	return &Node{value, hight, nil, nil, nil}
}

//插入树（把节点看成一个最简单的树）
func (node *Node) Insert(val int, high int) *Node {
	if node == nil {
		return CreateNode(val, high)
	} else {
		if val < node.value {
			node.left = node.Insert(val, high)
		} else {
			node.right = node.Insert(val, high)
		}
		return node
	}
}

//功能：查找节点，利用递归进行查找
//参数：根节点，查找的值
//返回值：查找值所在节点
func (node *Node) FindNode(n *Node, x int) *Node {
	if n == nil {
		return nil
	} else if n.value == x {
		return n
	} else {
		p := node.FindNode(n.left, x)
		if p != nil {
			return p
		}
		p = node.FindNode(n.mid, x)
		if p != nil {
			return p
		}
		return node.FindNode(n.right, x)
	}
}

//功能：求树的高度
//参数：根节点
//返回值：树的高度，树的高度=Max(左子树高度，右子树高度)+1
func (node *Node) GetTreeHeigh(n *Node) int {
	if n == nil {
		return 0
	} else {
		lHeigh := node.GetTreeHeigh(n.left)
		mHeigh := node.GetTreeHeigh(n.mid)
		rHeigh := node.GetTreeHeigh(n.right)
		var temp int
		if lHeigh > rHeigh {
			temp = lHeigh
			//return lHeigh+1
		} else {
			temp = rHeigh
		}
		if mHeigh > temp {
			return mHeigh + 1

		} else {
			return temp + 1
		}
	}
}

//功能：递归前序遍历二叉树
//参数：根节点
//返回值：nil
func (node *Node) PreOrder(n *Node) {
	if n != nil {
		fmt.Printf("(%d,%d)", n.value, n.hight)
		node.PreOrder(n.left)
		node.PreOrder(n.mid)
		node.PreOrder(n.right)
	}
}

type gametest struct {
	name     int
	height   int
	icondid  int32
	treeNode []*Node
	b        bool
}

func (g *gametest) PreOrder12(a []int32, n *Node, iconid int32, indexarr []int) {
	currenid := iconid
	if iconid == 0 {
		currenid = a[n.value]
	}
	fmt.Println("xia", currenid)
	indexarr = append(indexarr, n.value)
	bEnter := true
	if n.left != nil && (a[n.left.value] == 0 || a[n.left.value] == currenid || currenid == 0) {
		var tempindexarr []int
		tempindexarr = append(tempindexarr, indexarr...)
		g.PreOrder12(a, n.left, currenid, tempindexarr)
		bEnter = false
	}
	if n.mid != nil && (a[n.mid.value] == 0 || a[n.mid.value] == currenid || currenid == 0) {
		var tempindexarr []int
		tempindexarr = append(tempindexarr, indexarr...)
		g.PreOrder12(a, n.mid, currenid, tempindexarr)
		bEnter = false
	}
	if n.right != nil && (a[n.right.value] == 0 || a[n.right.value] == currenid || currenid == 0) {
		var tempindexarr []int
		tempindexarr = append(tempindexarr, indexarr...)
		g.PreOrder12(a, n.right, currenid, tempindexarr)
		bEnter = false
	}
	if bEnter {
		if n.hight > 2 {
			fmt.Println(indexarr)
			fmt.Println("=====", n.hight, n.value, currenid)
		}

		return
	}

}

//功能：递归中序遍历二叉树
//参数：根节点
//返回值：nil
func (node *Node) InOrder(n *Node) {

	if n != nil {
		node.InOrder(n.left)
		fmt.Printf("%d ", n.value)
		node.InOrder(n.right)
	}
}

//功能：递归后序遍历二叉树
//参数：根节点
//返回值：nil
func (node *Node) PostOrder(n *Node) {
	if n != nil {
		node.PostOrder(n.left)
		node.PostOrder(n.mid)
		node.PostOrder(n.right)
		fmt.Printf("%d ", n.value)
	}

}

//功能：打印所有的叶子节点
//参数：root
//返回值：nil
func (node *Node) GetLeafNode(n *Node) {
	if n != nil {
		if n.left == nil && n.right == nil {
			fmt.Printf(",%d ", n.value)
		}
		node.GetLeafNode(n.left)
		node.GetLeafNode(n.mid)
		node.GetLeafNode(n.right)
	}
}

func binaryTreePaths2(root *Node) []string {
	paths := []string{}

	var f func(t *Node, str string)
	f = func(t *Node, str string) {
		// 递归终止条件
		if t == nil {
			return
		}

		// 递归过程
		str = str + "->" + strconv.Itoa(t.value)
		if t.left == nil && t.right == nil && t.mid == nil {
			// 前两位是"->",根节点前面加"->"可以统一逻辑
			paths = append(paths, str[2:])
		}

		if t.left != nil {
			f(t.left, str)
		}
		if t.mid != nil {
			f(t.mid, str)
		}
		if t.right != nil {
			f(t.right, str)
		}
	}

	f(root, "")
	return paths
}
func binaryTreePaths3(root *Node) [][]int {
	//paths := []int{}
	var paths [][]int
	//var temp []int
	var f func(t *Node, path []int)
	f = func(t *Node, path []int) {
		// 递归终止条件
		if t == nil {
			return
		}

		// 递归过程
		path = append(path, t.value)
		//temp=append(temp,i)
		if t.left == nil && t.right == nil && t.mid == nil {
			// 前两位是"->",根节点前面加"->"可以统一逻辑
			paths = append(paths, path)

		}

		if t.left != nil {
			f(t.left, path)
		}
		if t.mid != nil {
			f(t.mid, path)
		}
		if t.right != nil {
			f(t.right, path)
		}
	}

	f(root, nil)
	return paths
}
func GetTreePath(root *Node) [][]int {
	//paths := []int{}
	var paths [][]int
	//var temp []int
	var f func(t *Node, path []int)
	f = func(t *Node, path []int) {
		// 递归终止条件
		if t == nil {
			return
		}

		// 递归过程
		path = append(path, t.value)
		//temp=append(temp,i)
		if t.left == nil && t.right == nil && t.mid == nil {
			// 前两位是"->",根节点前面加"->"可以统一逻辑
			paths = append(paths, path)
		}

		if t.left != nil {
			f(t.left, path)
		}
		if t.mid != nil {
			f(t.mid, path)
		}
		if t.right != nil {
			f(t.right, path)
		}
	}

	f(root, nil)
	return paths
}

func (g *Game) CreatTree1(cheatvalue int64) {
	var StageTree [5][5]*Node
	//var TreePath  [][][]int
	var result []*Node //当前查找结果值
	var temp []*Node   //中间值
	var Matrix []int32
	//从0-4 开始查找
	for q := 0; q <= 4; q++ {
		Matrix = getMatrixArr5(q)
		b := 0
		for j := 0; j <= 4; j++ {
			//值为-1跳过
			if Matrix[j] == -1 {
				continue
			}
			StageTree[q][j] = StageTree[q][j].Insert(j, 1)
			//新循当局结果初始化
			result = make([]*Node, 0)
			result = append(result, StageTree[q][j])
			b++
			//Allresult=append(Allresult,t)
			//查找节点值
			//遍历后面4列数据

			for i := 0; i < 4; i++ {
				var first, fi, fj, last, li, lj, mi, mj int
				//每列循环 头尾 中间数字找的相邻位置，5个阶段每个阶段相邻的不一样
				if q == 0 {
					switch i {
					case 0:
						first = 1
						fi = 5
						fj = 6
						last = 3
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 1:
						first = 6
						fi = 5
						fj = 6
						last = 8
						li = 5
						lj = 6
						mi = 5
						mj = 6
					case 2:
						first = 11
						fi = 5
						fj = 6
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 3:
						first = 16
						fi = 4
						fj = 5
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 5
					}
				} else if q == 1 {
					switch i {
					case 0:
						first = 1
						fi = 5
						fj = 6
						last = 3
						li = 5
						lj = 6
						mi = 4
						mj = 6
					case 1:
						first = 6
						fi = 5
						fj = 6
						last = 9
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 2:
						first = 11
						fi = 4
						fj = 5
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 5
					case 3:
						first = 15
						fi = 5
						fj = 6
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 6
					}

				} else if q == 2 {
					switch i {
					case 0:
						first = 1
						fi = 5
						fj = 6
						last = 4
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 1:
						first = 6
						fi = 4
						fj = 5
						last = 9
						li = 4
						lj = 5
						mi = 4
						mj = 5
					case 2:
						first = 10
						fi = 5
						fj = 6
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 3:
						first = 15
						fi = 5
						fj = 6
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 6
					}
				} else if q == 3 {
					switch i {
					case 0:
						first = 1
						fi = 4
						fj = 5
						last = 4
						li = 4
						lj = 5
						mi = 4
						mj = 5
					case 1:
						first = 5
						fi = 5
						fj = 6
						last = 9
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 2:
						first = 10
						fi = 5
						fj = 6
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 3:
						first = 15
						fi = 5
						fj = 6
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 6
					}
				} else if q == 4 {
					switch i {
					case 0:
						first = 0
						fi = 5
						fj = 6
						last = 4
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 1:
						first = 5
						fi = 5
						fj = 6
						last = 9
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 2:
						first = 10
						fi = 5
						fj = 6
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 3:
						first = 15
						fi = 5
						fj = 6
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 6
					}
				}
				//上局结果赋值给中间变量
				temp = result
				//初始化当局结果
				result = make([]*Node, 0)
				//上局是否有结果 if有则继续下一列查找 else 退出本次查找
				//遍历上局结果，用每个结果便利下一下列查找是否有相邻的
				for _, v := range temp {
					index := 0 //赋值到树的下表，012 对应Left mid right
					var tempi, tempj int
					if v.value == first {
						tempi = fi
						tempj = fj
					} else if v.value == last {
						tempi = li
						tempj = lj
					} else {
						tempi = mi
						tempj = mj
					}
					//每个值需要查找相邻的下标循环
					for k := tempi; k <= tempj; k++ {
						tr := v
						switch index {
						case 0:
							tr.left = CreateNode(v.value+k, i+2)
							result = append(result, tr.left)
						case 1:
							tr.mid = CreateNode(v.value+k, i+2)
							result = append(result, tr.mid)
						case 2:
							tr.right = CreateNode(v.value+k, i+2)
							result = append(result, tr.right)
						}
						index++
					}
				}
			}
		}
	}

	//StageTree[0][1].PreOrder(StageTree[1][2])
	var ggg gametest
	//a := []int32{-1, 4, 8, 7, -1,
	//	         -1, 9, 0, 4, -1,
	//	         -1, 8, 8, 5, 7,
	//	          -4, 4, 8, 4, 7,
	//	          7, 3, 9, 1, 4}
	a := []int32{-1, 3, 0, 7, -1,
		-1, 2, 8, 2, 4,
		-1, 0, 8, 3, 8,
		5, 8, 3, 4, 8,
		5, 4, 0, 7, 4}
	StageTree[1][2].PreOrder(StageTree[1][2])
	for i := 1; i <= 3; i++ {
		var aa []int

		ggg.PreOrder12(a, StageTree[1][i], a[StageTree[1][i].value], aa)
	}
	//ggg.PreOrder12(a, StageTree[1][2], a[StageTree[1][2].value])
	//fmt.Println(binaryTreePaths3(StageTree[0][1]))
	//for i:=0;i<len(StageTree);i++ {
	//	//fmt.Println("-------阶段%v------", i)
	//	for j := 0; j < len(StageTree[i]); j++ {
	//		fmt.Println("--- 树%v%v---", i, j)
	//		fmt.Println(binaryTreePaths3(StageTree[i][j]))
	//		TreePath= append(TreePath,binaryTreePaths3(StageTree[i][j]))
	//
	//	}
	//}

	return

}
