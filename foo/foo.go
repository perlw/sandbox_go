package foo

// void foo(void);
import "C"

func Foo() {
	C.foo()
}
