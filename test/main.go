package main

import "fmt"

func main() {
	for i := 0; i < 5; i++ {
		for j := 0; j < 4; j++ {
			if j == 2 {
				break
			}
			fmt.Print(j)
		}
		fmt.Println()
	}
}
