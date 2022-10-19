package internal

import (
	"reflect"
	"unsafe"
)

func StringToBytes(s string) (b []byte) {
	sHdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bHdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bHdr.Data = sHdr.Data
	bHdr.Len = sHdr.Len
	bHdr.Cap = sHdr.Len
	return
}
