// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved

package qf

import (
	"encoding/binary"
	"io"
	"reflect"
	"unsafe"
)

var isLittleEndian bool

func init() {
	buf := []byte{0x1, 0x0}
	val := (*uint16)(unsafe.Pointer((*(*reflect.SliceHeader)(unsafe.Pointer(&buf))).Data))
	isLittleEndian = *val == uint16(1)
}

func unsafeUint64SliceToBytes(space []uint64) []byte {
	// Get the slice header
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&space))

	// The length and capacity of the slice are different.
	header.Len *= bytesPerWord
	header.Cap *= bytesPerWord

	// Convert slice header to an []byte
	return *(*[]byte)(unsafe.Pointer(&header))
}

func writeUintSlice(w io.Writer, v []uint64) (n int64, err error) {
	if err = binary.Write(w, binary.LittleEndian, uint64(len(v))); err != nil {
		return
	}
	n += 8
	if isLittleEndian {
		// ~12x faster
		data := unsafeUint64SliceToBytes(v)
		var np int
		np, err = w.Write(data)
		n += int64(np)
	} else {
		err = binary.Write(w, binary.LittleEndian, v)
		if err == nil {
			n += int64(len(v)) * 8
		}
	}
	return
}

func readUintSlice(r io.Reader) (v []uint64, n int64, err error) {
	// read length
	var length uint64
	err = binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return
	}
	n += 8
	v = make(unpacked, length)
	if isLittleEndian {
		// ~15x faster
		data := unsafeUint64SliceToBytes(v)
		var np int
		np, err = r.Read(data)
		n += int64(np)
	} else {
		err = binary.Read(r, binary.LittleEndian, v)
		if err != nil {
			return
		}
		n += 8 * int64(length)
	}
	return
}
