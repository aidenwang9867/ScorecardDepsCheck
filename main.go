package main

import (
	"fmt"
)

func main() {
	for i := 0; i < 5; i++ {
		fmt.Println("hello world" + fmt.Sprintf("%v", i))
	}
	return
}
