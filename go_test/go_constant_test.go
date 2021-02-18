package go_test

import (
	"fmt"
	"testing"
)

const (
	Read = 1 << iota
	Write
	Execute
)

func TestConstant(t *testing.T) {

	t.Log("hello", Read, Write, Execute)

	t.Log(Read &^ 1)
	t.Log(Write &^ 1)
	t.Log(7 &^ 1)
}

func TestMap(t *testing.T) {
	m := map[string]string{}
	var mm = make(map[string]string, 2)

	m["a"] = "A"
	mm["a"] = "A"

	t.Log(m["a"])
	t.Log(mm["a"])

}

func TestArray(t *testing.T) {
	aa := [...]int{}
	bb := []int{}
	t.Logf("%T", aa)
	t.Logf("%T", bb)
}

type Program interface {
	Echo() string
}

func ExecuteEcho(p Program) {
	fmt.Printf("%T,%s", p, p.Echo())
}

type GoProgram struct {
}

func (p *GoProgram) Echo() string {
	return "fmt.Println(\"Hello World\")\n"
}

type JavaProgram struct {
}

func (p *JavaProgram) Echo() string {
	return "System.out.println(\"Hello World\")\n"
}

func TestPolymorphism(t *testing.T) {
	goProgram := new(GoProgram)
	javaProgram := new(JavaProgram)

	ExecuteEcho(goProgram)
	ExecuteEcho(javaProgram)
}
