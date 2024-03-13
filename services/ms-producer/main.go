package main

import (
	"fmt"

	"github.com/ifeel3/RieltaExercise/pkg/structs"
)

func main() {
	fmt.Println("Hello from producer")
	usr := structs.User{Id: 1, Name: "Denis", Balance: 100}
	fmt.Println(usr)
}
