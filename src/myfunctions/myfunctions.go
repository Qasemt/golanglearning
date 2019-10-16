package myfunctions

import (
	"fmt"
	"time"
)

func SayHello() string {
	return "Hello from this another package"
}

func TestSwitch() {
	t := time.Now()
	switch {
	case t.Hour() < 12:
		fmt.Println("It's before noon")
	default:
		fmt.Println("It's after noon")
	}
}

func Testarray() {
	var a [3]int //int array with length 3
	a[0] = 12    // array index starts at 0
	a[1] = 78
	a[2] = 50
	fmt.Println(a)
	b := [3]int{22, 7448, 4} // short hand declaration to create array
	fmt.Println(b)
	c := [...]int{12, 78, 50} // ... makes the compiler determine the length
	fmt.Println(c)

	// ::::::::::::::::::::::::::::::::::::::::::::::
	j := [5]int{44, 12, 33, 786, 80}
	var t []int = j[0:4] //creates a slice from a[1] to a[3]
	fmt.Println(t)
	// ::::::::::::::::::::::::::::::::::::::::::::::
	darr := [...]int{57, 89, 90, 82, 100, 78, 67, 69, 59}
	dslice := darr[2:5]
	fmt.Println("array before", darr)
	for i := range dslice {
		dslice[i]++
	}
	fmt.Println("array after", darr)
}
