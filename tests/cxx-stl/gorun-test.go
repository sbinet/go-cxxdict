package main

import (
	"fmt"

	// 3rd-party
	"mylib"
)

func main() {

	fmt.Printf("mylib.NewClass()...\n")
	c := mylib.NewClass()
	fmt.Printf("mylib.NewClass()... [ok]\n")

	fmt.Printf("mylib.DeleteClass(c)...\n")
	mylib.DeleteClass(c)
	fmt.Printf("mylib.DeleteClass(c)... [ok]\n")

	fmt.Printf("mylib.NewNamed(\"n\")...\n")
	n := mylib.NewNamed("n")
	fmt.Printf("mylib.NewNamed(\"n\")... [ok]\n")

	nn := n.Name()
	fmt.Printf("n.Name() = \"%s\"\n", nn)

	nn = "boo"
	fmt.Printf("n.SetName(\"%s\"\n", nn)
	n.SetName(nn)

	nn = n.Name()
	fmt.Printf("n.Name() = \"%s\"\n", nn)

	fmt.Printf("mylib.DeleteNamed(n)...\n")
	mylib.DeleteNamed(n)
	fmt.Printf("mylib.DeleteNamed(n)... [ok]\n")
}
