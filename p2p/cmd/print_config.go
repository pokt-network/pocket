package main

import "fmt"

func main() {

	fmt.Println("version: \"3.5\"")
	fmt.Println("services:")
	for i := 0; i < 27; i++ {
		fmt.Printf("\n\n  node.%d:\n    image: p2p\n    build:\n      dockerfile:./build/docker/node.Dockerfile\n    environment:\n      - ADDRESS=0.0.0.0:120%d\n    ports:\n      - \"120%d:120%d\"", i+1, i+1, i+1, i+1)
	}
}
