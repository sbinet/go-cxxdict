package main

import (
	"fmt"

	// 3rd-party
	"mylib"
)

func main() {
	
	fmt.Printf("gocxx_mylib.Add42(100)= %v\n",
		mylib.Add42(100))

	fmt.Printf("gocxx_mylib.Add42(0.1)= %v\n",
		mylib.Add42(0.1))

	fmt.Printf("creating a new Foo...\n")
	foo := mylib.NewFoo()
	fmt.Printf("foo.GetDouble()= %v\n", foo.GetDouble())
	fmt.Printf("foo.SetDouble(42)...\n")
	foo.SetDouble(42)
	fmt.Printf("foo.GetDouble()= %v\n", foo.GetDouble())
	fmt.Printf("foo.SetDouble(.4)...\n")
	foo.SetDouble(0.4)
	fmt.Printf("foo.GetDouble()= %v\n", foo.GetDouble())
	fmt.Printf("deleting foo...\n")
	mylib.DeleteFoo(foo)
	fmt.Printf("deleting foo...[done]\n")
}
