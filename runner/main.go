package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("starting runner...")
	go forever()
	select {} // block forever
}

func forever() {
	fmt.Println("started runner")
	for {
		fmt.Printf("%d\n", time.Now().Unix())
		time.Sleep(time.Second * 60)
	}
}
