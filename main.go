package main

import "demo/calculator"
import "fmt"

func main() {
	a := 1
	b := -5
	sum, isNeg := calculator.AddAndCheckNegative(a, b) 
	fmt.Println(sum, isNeg)
}
