package util

import (
	"fmt"
	"testing"
)

func Test_List(t *testing.T) {
	list := NewList(10, "a", "b", "c")
	fmt.Println("list.Size():", list.Size())
	fmt.Println("list.Gets():", list.Gets())
	list.Add("a1")
	list.Adds("b1", "b2")
	fmt.Println("list.Gets():", list.Gets())
	fmt.Println("list.Get(1):", list.Get(1))
	fmt.Println("list.GetIndex(c):", list.GetIndex("c"))
	list.Insert(0, "z0")
	fmt.Println("list.Gets():", list.Gets())
	list.Insert(2, "z2")
	fmt.Println("list.Gets():", list.Gets())
	list.Insert(list.Size(), "zend")
	fmt.Println("list.Gets():", list.Gets())
	list.Remove(1)
	fmt.Println("list.Gets():", list.Gets())
}
