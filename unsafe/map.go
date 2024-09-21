package unsafe

import (
	"hash/maphash"
	"unsafe"
)

func UnsafeHashOf[T comparable]() func(seed maphash.Seed, v T) uint64 {
	h := getHashFunc[T](map[T]struct{}{})
	return func(seed maphash.Seed, v T) uint64 {
		return h.Sum64(seed, v)
	}
}

// https://github.com/golang/go/blob/master/src/internal/abi/type.go#L20
type abiType struct {
	Size_       uintptr
	PtrBytes    uintptr // number of (prefix) bytes in the type that can contain pointers
	Hash        uint32  // hash of type; avoids computation in hash tables
	TFlag       uint8   // extra type information flags
	Align_      uint8   // alignment of variable with this type
	FieldAlign_ uint8   // alignment of struct field with this type
	Kind_       uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	// GCData stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, GCData is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	GCData    *byte
	Str       int32 // string form
	PtrToThis int32 // type for pointer to this type, may be zero
}

// https://github.com/golang/go/blob/master/src/internal/abi/map_swiss.go#L25
type abiMapType[T comparable] struct {
	abiType
	Key    *abiType
	Elem   *abiType
	Bucket *abiType // internal type representing a hash bucket
	// function for hashing keys (ptr to key, seed) -> hash
	Hasher     hashFunc[T]
	KeySize    uint8  // size of key slot
	ValueSize  uint8  // size of elem slot
	BucketSize uint16 // size of bucket
	Flags      uint32
}

type hashFunc[T comparable] func(value *T, seed maphash.Seed) uintptr

func (h hashFunc[T]) Sum64(seed maphash.Seed, v T) uint64 { return uint64(h(&v, seed)) }

type emptyInterface struct{ typ *abiType }

func getHashFunc[T comparable](v interface{}) hashFunc[T] {
	return (*abiMapType[T])(unsafe.Pointer((*emptyInterface)(unsafe.Pointer(&v)).typ)).Hasher
}
