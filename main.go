package main

import (
	"fmt"
	"os"

	"github.com/cubeflix/edit/editor"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: ke file")
	} else if len(os.Args) == 2 {
		e, err := editor.NewEditor(os.Args[1])
		if err != nil {
			panic(err)
		}
		err = e.Init()
		if err != nil {
			panic(err)
		}
		e.HandleEvents()
	} else {
		fmt.Println("usage: ke file")
	}
}
