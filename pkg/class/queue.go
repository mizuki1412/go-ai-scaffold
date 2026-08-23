package class

import "sync"

type (
	Queue[T any] struct {
		top    *node[T]
		rear   *node[T]
		length int
		lock1  sync.Mutex
		lock2  sync.Mutex
	}
	//双向链表节点
	node[T any] struct {
		pre   *node[T]
		next  *node[T]
		value T
	}
)

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (th *Queue[T]) Len() int {
	return th.length
}

func (th *Queue[T]) Peek() *T {
	if th.top == nil {
		return nil
	}
	return &th.top.value
}

func (th *Queue[T]) Push(v T) {
	th.lock1.Lock()
	defer th.lock1.Unlock()
	n := &node[T]{nil, nil, v}
	if th.length == 0 {
		th.top = n
		th.rear = th.top
	} else {
		n.pre = th.rear
		th.rear.next = n
		th.rear = n
	}
	th.length++
}

func (th *Queue[T]) Pop() *T {
	th.lock2.Lock()
	defer th.lock2.Unlock()
	if th.length == 0 {
		return nil
	}
	n := th.top
	if th.top.next == nil {
		th.top = nil
	} else {
		th.top = th.top.next
		th.top.pre.next = nil
		th.top.pre = nil
	}
	th.length--
	return &n.value
}
