package testFunctions

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func sayHello() string {
	return "Hello from this another package"
}

func testSwitch() {
	t := time.Now()
	switch {
	case t.Hour() < 12:
		fmt.Println("It's before noon")
	default:
		fmt.Println("It's after noon")
	}
}

func testarray() {
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

func testtime() {

	epoch := time.Now().Unix()
	fmt.Println(epoch)
}

func testTcpServer() {
	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, _ := net.Listen("tcp", ":4444")

	// accept connection on port
	conn, _ := ln.Accept()

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, _ := bufio.NewReader(conn).ReadString('\n')
		// output message received
		fmt.Print("Message Received:", string(message))
		// sample process for string received
		newmessage := strings.ToUpper(message)
		// send new string back to client
		conn.Write([]byte("\n rec from server : " + newmessage + "\n"))
	}
}
func say(s string) {
	for i := 0; i < 1000; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}
func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c
}

func squares(c chan int) {

	for i := 0; i <= 3; i++ {
		n := <-c
		fmt.Println(n * n)
	}

}
func testThread() {
	go say("world")
	say("hello")
}

func testThreadChannal() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)
}
func testChannalBuffer() {

	c := make(chan int, 3)
	go squares(c)
	c <- 1
	c <- 2
	c <- 3

}
func RunTest() {

	// if testCSV() {
	// 	fmt.Printf("csv export success \n")
	// }
	//testTcpServer()
	//:::::::::::::::::::::::::: THREAD
	//testThread()
	testThreadChannal()
	//	testChannalBuffer()
}
