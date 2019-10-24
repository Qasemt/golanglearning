package testFunctions

import (
	"fmt"
	"time"
)

func RunTestGoRoot() {
	tick := time.Tick(1000 * time.Millisecond)
	boom := time.After(5000 * time.Millisecond)
	for {
		select {
		case <-tick:
			fmt.Println("tick.")
		case <-boom:
			fmt.Println("BOOM!")
			return
		default:
			//	fmt.Println("    .")
			//	time.Sleep(50 * time.Millisecond)
		}
	}
}
