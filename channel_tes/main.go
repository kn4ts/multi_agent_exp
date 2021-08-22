package main

import (
	"fmt"
	"time"
)

type MyStruct struct{
	Mss	*MySubStruct
}

type MySubStruct struct{
	ch	chan int
}

func NewMyStruct() *MyStruct{
	mss := NewMySubStruct()
	return &MyStruct {
		Mss: mss,
	}
}

func NewMySubStruct() *MySubStruct {
	return &MySubStruct{
		ch: make(chan int),
	}
}

func ( mst MyStruct ) SendFunc() {
	var i int
	for {
		mst.Mss.ch <- 1
		fmt.Println("   ",i)
		time.Sleep(2*time.Second)
		i++
	}
}

func ( mss MySubStruct) RecieveFunc() {
	var v int
	for {
		select {
		case v = <- mss.ch:
		default:
			v = 0
		}
		fmt.Println(v, len(mss.ch))
		time.Sleep(time.Second)
	}
}

func main() {

	st := NewMyStruct()
	//st := MyStruct{}
	//st.ch = make(chan int)
	
	go st.SendFunc()

	go st.Mss.RecieveFunc()

	for{}
}
