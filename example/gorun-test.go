package main

// stdlib imports
import (
	"fmt"
)

// 3rd party imports
import (
	foo "mylibpkg"
)

func test_panic_doadd() {

	fmt.Printf("--- testing panic ---\n")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in test_panic_doadd:", r)
		}
	}()

	fmt.Printf("do_add(\"x\")= %v\n", foo.Math_do_add("x"))
}

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main:", r)
		}
	}()

	fmt.Printf("\n\n:::::::::::::::::::::::::::::::\n")
	fmt.Printf("Math::do_hello()...\n")
	foo.Math_do_hello()
	fmt.Printf("Math2::do_hello(\"go:\")...[%s]\n", foo.Math2_do_hello("go:"))
	fmt.Printf("Math2::do_hello()...       [%s]\n", foo.Math2_do_hello())

	fmt.Printf("Math2::do_hello_c(\"go:\")...[%s]\n", foo.Math2_do_hello_c("go:"))

	fmt.Printf("Math::say_hello()...\n")
	fmt.Printf("go: %s\n", foo.Math_say_hello())

	fmt.Printf("adder(41,42)=       %v\n", foo.Math_adder(41, 42))
	fmt.Printf("do_add(  )=         %v\n", foo.Math_do_add())
	fmt.Printf("do_add(1.)=         %v\n", foo.Math_do_add(1.))
	fmt.Printf("do_add(1 )=         %v\n", foo.Math_do_add(1))
	fmt.Printf("do_add(1.,2.)=      %v\n", foo.Math_do_add(1., 2.))
	fmt.Printf("do_add(1, 2)=       %v\n", foo.Math_do_add(1, 2))
	fmt.Printf("do_add(1, 2)=       %v\n", foo.Math_do_add(1, 2))
	fmt.Printf("do_add(1, 2, 3)=    %v\n", foo.Math_do_add(1, 2, 3))
	fmt.Printf("do_add(1, 2, 3, 4)= %v\n", foo.Math_do_add(1, 2, 3, 4))

	test_panic_doadd()

	fmt.Printf("::bye.\n")
}
