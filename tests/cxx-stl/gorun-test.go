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

	fmt.Printf("c.Nbr_ints()= ?\n")
	fmt.Printf("c.Nbr_ints()= %v\n", c.Nbr_ints())

	fmt.Printf("c.Add(42)...\n")
	c.Add(42)
	fmt.Printf("c.Nbr_ints()= ?\n")
	fmt.Printf("c.Nbr_ints()= %v\n", c.Nbr_ints())

	fmt.Printf("c.Nbr_doubles()= ?\n")
	fmt.Printf("c.Nbr_doubles()= %v\n", c.Nbr_doubles())

	fmt.Printf("c.Add(0.66)...\n")
	c.Add(0.66)
	fmt.Printf("c.Nbr_doubles()= ?\n")
	fmt.Printf("c.Nbr_doubles()= %v\n", c.Nbr_doubles())

	fmt.Printf("mylib.DeleteClass(c)...\n")
	mylib.DeleteClass(c)
	fmt.Printf("mylib.DeleteClass(c)... [ok]\n")

	fmt.Printf("mylib.NewNamed(\"n\")...\n")
	n := mylib.NewNamed("n")
	fmt.Printf("mylib.NewNamed(\"n\")... [ok]\n")

	fmt.Printf("n.Name() = ???\n")
	nn := n.Name()
	fmt.Printf("n.Name() = \"%s\"\n", nn)

	nn = "boo"
	fmt.Printf("n.SetName(\"%s\")\n", nn)
	n.SetName(nn)

	nn = n.Name()
	fmt.Printf("n.Name() = \"%s\"\n", nn)

	fmt.Printf("mylib.DeleteNamed(n)...\n")
	mylib.DeleteNamed(n)
	fmt.Printf("mylib.DeleteNamed(n)... [ok]\n")
}
