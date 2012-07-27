package main

import (
	"fmt"
	"os"

	// 3rd-party
	"mylib"
)

func main() {

	ids := []int{0,1,2,3,4}
	fmt.Printf("mylib.NewApp()...\n")
	app := mylib.NewApp()
	fmt.Printf("mylib.NewApp()... [ok]\n")

	for _,id := range ids {
		n := fmt.Sprintf("alg_%d", id)
		fmt.Printf("mylib.NewAlg(\"%s\")...\n", n)
		alg := mylib.NewAlg(n)
		fmt.Printf("mylib.NewAlg(\"%s\")...[ok]\n", alg.Name())

		fmt.Printf("adding alg to app...\n")
		sc := app.AddAlg(alg)
		if sc != 0 {
			fmt.Printf("adding alg to app...[err]\n")
			os.Exit(1)
		}
		fmt.Printf("adding alg to app...[ok]\n")
		
	}

	fmt.Printf("running app...\n")
	sc := app.Run()
	if sc != 0 {
		fmt.Printf("running app...[err]\n")
		os.Exit(1)
	}
	fmt.Printf("running app...[ok]\n")

	
	fmt.Printf("deleting app...\n")
	mylib.DeleteApp(app)
	fmt.Printf("deleting app...[ok]\n")
}
