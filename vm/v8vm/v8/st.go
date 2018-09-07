package main

/*
#include <stdlib.h>
#include "vm.h"
#cgo darwin LDFLAGS: -lvm
#cgo linux LDFLAGS: -L${SRCDIR}/libv8/_linux_amd64 -lvm -lv8 -Wl,-rpath,${SRCDIR}/libv8/_linux_amd64
*/
import "C"
import "fmt"

func main() {
	C.init()
	a := C.createStartupData()
	fmt.Println(a)
}
