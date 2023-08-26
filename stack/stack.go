package stack

type Stack []interface{}

func (s *Stack) Push(item interface{}) {
	*s = append(*s, item)
}

func (s *Stack) Pop() interface{} {
	if len(*s) == 0 {
		return nil
	}
	item := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return item
}

// Peek 返回栈顶元素，但不弹出
func (s *Stack) Peek() interface{} {
	if len(*s) == 0 {
		return nil
	}
	return (*s)[len(*s)-1]
}

func (s *Stack) IsEmpty() bool {
	return len(*s) == 0
}
