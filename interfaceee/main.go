package main

import "fmt"

type InterfaceX interface {
	Sum(a, b int) (int, error)
	Multi(a, b int) (int, error)
} 


type S struct {
	f1 int 
}

func (s S) Sum(a, b int) (int, error) { // method: -> receiver: value / pointer
	return a+b, nil 
}

func (s S) Multi(a, b int) (int, error) { // method: -> receiver: value / pointer
	return a*b, nil 
}


type S2 struct {

}

func (s2 S2) Sum(a, b int) (int, error) {
	return -(a+b), nil 
}

func (s2 S2) Multi(a, b int) (int, error) { 
	return a*b, nil 
}

func main() {
	// interface's value: (underlying type, underlying value)
	// var i InterfaceX  // value = (nil, nil)
	
	// var s S // S{f1: 0}
	
	// i = s // value = (S, s)

	// sum, _ := i.Sum(1,1)

	// fmt.Println(sum)

	// var s2 S2 = S2{}
	// i = s2 
	// sum2, _ := s2.Sum(2,2)

	// fmt.Println(sum2)

	// // check interface underlying type : type assertion
	// underlying, ok := i.(S) 

	// fmt.Printf("ok = %v, type interface = %T\n", ok, underlying)

	// switch i.(type) {
	// case S:
	// 	fmt.Println("current underlying type is S")
	// case S2:
	// 	fmt.Println("current underlying type is S2")
	// default:
	// 	fmt.Println("undetected type")
	// }

	// // empty interface: interface have no funcs 
	// var emptyI interface{} // -> all types implement empty interface 

	// emptyI = s

	// fmt.Printf("type=%T, value=%+v\n", emptyI, emptyI) // S, S{f1: 10}

	// // nil interface: interface is nil <=> (underlying type = nil, underlying value = nil)

	// var nilI InterfaceX

	// fmt.Printf("nilI is nil ? %v\n", nilI == nil)

	// var s3 *S3 = nil 
	// nilI = s3 // (underlying type = *S3, underlying value = nil)

	// fmt.Printf("nilI is nil ? %v\n", nilI == nil) // false 

	// check if a type implements an interface 
	var _ InterfaceX = S{}
	var _ InterfaceX = S2{}
	var _ InterfaceX = &S3{} // struct // 

	fmt.Println("ok")
}

type S3 struct {

}

func (s *S3) Sum(a, b int) (int, error) {
	return a+b, nil 
} 

func (s *S3) Multi(a, b int) (int, error) {
	return a*b, nil 
} 