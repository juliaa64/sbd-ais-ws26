package main

import (
	"fmt"
)

// struct
type Saludo struct {
	Mensaje string
}

// reciver
func (s *Saludo) Imprimir() {
	fmt.Printf(s.Mensaje)
}

func main() {
	s := Saludo{Mensaje: "Hello World!"}
	s.Imprimir()
}
