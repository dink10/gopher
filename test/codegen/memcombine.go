// asmcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import (
	"encoding/binary"
	"runtime"
)

var sink64 uint64
var sink32 uint32
var sink16 uint16

// ------------- //
//    Loading    //
// ------------- //

func load_le64(b []byte) {
	// amd64:`MOVQ\s\(.*\),`
	// s390x:`MOVDBR\s\(.*\),`
	// arm64:`MOVD\s\(R[0-9]+\),`,-`MOV[BHW]`
	// ppc64le:`MOVD\s`,-`MOV[BHW]Z`
	sink64 = binary.LittleEndian.Uint64(b)
}

func load_le64_idx(b []byte, idx int) {
	// amd64:`MOVQ\s\(.*\)\(.*\*1\),`
	// s390x:`MOVDBR\s\(.*\)\(.*\*1\),`
	// arm64:`MOVD\s\(R[0-9]+\)\(R[0-9]+\),`,-`MOV[BHW]`
	// ppc64le:`MOVD\s`,-`MOV[BHW]Z\s`
	sink64 = binary.LittleEndian.Uint64(b[idx:])
}

func load_le32(b []byte) {
	// amd64:`MOVL\s\(.*\),`           386:`MOVL\s\(.*\),`
	// s390x:`MOVWBR\s\(.*\),`
	// arm64:`MOVWU\s\(R[0-9]+\),`,-`MOV[BH]`
	// ppc64le:`MOVWZ\s`
	sink32 = binary.LittleEndian.Uint32(b)
}

func load_le32_idx(b []byte, idx int) {
	// amd64:`MOVL\s\(.*\)\(.*\*1\),`  386:`MOVL\s\(.*\)\(.*\*1\),`
	// s390x:`MOVWBR\s\(.*\)\(.*\*1\),`
	// arm64:`MOVWU\s\(R[0-9]+\)\(R[0-9]+\),`,-`MOV[BH]`
	// ppc64le:`MOVWZ\s`
	sink32 = binary.LittleEndian.Uint32(b[idx:])
}

func load_le16(b []byte) {
	// amd64:`MOVWLZX\s\(.*\),`
	// ppc64le:`MOVHZ\s`
	// arm64:`MOVHU\s\(R[0-9]+\),`,-`MOVB`
	sink16 = binary.LittleEndian.Uint16(b)
}

func load_le16_idx(b []byte, idx int) {
	// amd64:`MOVWLZX\s\(.*\),`
	// ppc64le:`MOVHZ\s`
	// arm64:`MOVHU\s\(R[0-9]+\)\(R[0-9]+\),`,-`MOVB`
	sink16 = binary.LittleEndian.Uint16(b[idx:])
}

func load_be64(b []byte) {
	// amd64:`BSWAPQ`
	// s390x:`MOVD\s\(.*\),`
	// arm64:`REV`,`MOVD\s\(R[0-9]+\),`,-`MOV[BHW]`,-`REVW`,-`REV16W`
	sink64 = binary.BigEndian.Uint64(b)
}

func load_be64_idx(b []byte, idx int) {
	// amd64:`BSWAPQ`
	// s390x:`MOVD\s\(.*\)\(.*\*1\),`
	// arm64:`REV`,`MOVD\s\(R[0-9]+\)\(R[0-9]+\),`,-`MOV[WHB]`,-`REVW`,-`REV16W`
	sink64 = binary.BigEndian.Uint64(b[idx:])
}

func load_be32(b []byte) {
	// amd64:`BSWAPL`
	// s390x:`MOVWZ\s\(.*\),`
	// arm64:`REVW`,`MOVWU\s\(R[0-9]+\),`,-`MOV[BH]`,-`REV16W`
	sink32 = binary.BigEndian.Uint32(b)
}

func load_be32_idx(b []byte, idx int) {
	// amd64:`BSWAPL`
	// s390x:`MOVWZ\s\(.*\)\(.*\*1\),`
	// arm64:`REVW`,`MOVWU\s\(R[0-9]+\)\(R[0-9]+\),`,-`MOV[HB]`,-`REV16W`
	sink32 = binary.BigEndian.Uint32(b[idx:])
}

func load_be16(b []byte) {
	// amd64:`ROLW\s\$8`
	// arm64: `REV16W`,`MOVHU\s\(R[0-9]+\),`,-`MOVB`
	sink16 = binary.BigEndian.Uint16(b)
}

func load_be16_idx(b []byte, idx int) {
	// amd64:`ROLW\s\$8`
	// arm64: `REV16W`,`MOVHU\s\(R[0-9]+\)\(R[0-9]+\),`,-`MOVB`
	sink16 = binary.BigEndian.Uint16(b[idx:])
}

func load_byte2_uint16(s []byte) uint16 {
	// arm64:`MOVHU\t\(R[0-9]+\)`,-`ORR\tR[0-9]+<<8`
	return uint16(s[0]) | uint16(s[1])<<8
}

// Check load combining across function calls.

func fcall_byte(a, b byte) (byte, byte) {
	return fcall_byte(fcall_byte(a, b)) // amd64:`MOVW`
}

func fcall_uint16(a, b uint16) (uint16, uint16) {
	return fcall_uint16(fcall_uint16(a, b)) // amd64:`MOVL`
}

func fcall_uint32(a, b uint32) (uint32, uint32) {
	return fcall_uint32(fcall_uint32(a, b)) // amd64:`MOVQ`
}

// We want to merge load+op in the first function, but not in the
// second. See Issue 19595.
func load_op_merge(p, q *int) {
	x := *p
	*q += x // amd64:`ADDQ\t\(`
}
func load_op_no_merge(p, q *int) {
	x := *p
	for i := 0; i < 10; i++ {
		*q += x // amd64:`ADDQ\t[A-Z]`
	}
}

// Make sure offsets are folded into loads and stores.
func offsets_fold(_, a [20]byte) (b [20]byte) {
	// arm64:`MOVD\t""\.a\+[0-9]+\(FP\), R[0-9]+`,`MOVD\tR[0-9]+, ""\.b\+[0-9]+\(FP\)`
	b = a
	return
}

// Make sure we don't put pointers in SSE registers across safe
// points.

func safe_point(p, q *[2]*int) {
	a, b := p[0], p[1] // amd64:-`MOVUPS`
	runtime.GC()
	q[0], q[1] = a, b // amd64:-`MOVUPS`
}

// ------------- //
//    Storing    //
// ------------- //

func store_le64(b []byte) {
	// amd64:`MOVQ\s.*\(.*\)$`,-`SHR.`
	// arm64:`MOVD`,-`MOV[WBH]`
	// ppc64le:`MOVD\s`,-`MOV[BHW]\s`
	binary.LittleEndian.PutUint64(b, sink64)
}

func store_le64_idx(b []byte, idx int) {
	// amd64:`MOVQ\s.*\(.*\)\(.*\*1\)$`,-`SHR.`
	// arm64:`MOVD\sR[0-9]+,\s\(R[0-9]+\)\(R[0-9]+\)`,-`MOV[BHW]`
	// ppc64le:`MOVD\s`,-`MOV[BHW]\s`
	binary.LittleEndian.PutUint64(b[idx:], sink64)
}

func store_le32(b []byte) {
	// amd64:`MOVL\s`
	// arm64:`MOVW`,-`MOV[BH]`
	// ppc64le:`MOVW\s`
	binary.LittleEndian.PutUint32(b, sink32)
}

func store_le32_idx(b []byte, idx int) {
	// amd64:`MOVL\s`
	// arm64:`MOVW\sR[0-9]+,\s\(R[0-9]+\)\(R[0-9]+\)`,-`MOV[BH]`
	// ppc64le:`MOVW\s`
	binary.LittleEndian.PutUint32(b[idx:], sink32)
}

func store_le16(b []byte) {
	// amd64:`MOVW\s`
	// arm64:`MOVH`,-`MOVB`
	// ppc64le(DISABLED):`MOVH\s`
	binary.LittleEndian.PutUint16(b, sink16)
}

func store_le16_idx(b []byte, idx int) {
	// amd64:`MOVW\s`
	// arm64:`MOVH\sR[0-9]+,\s\(R[0-9]+\)\(R[0-9]+\)`,-`MOVB`
	// ppc64le(DISABLED):`MOVH\s`
	binary.LittleEndian.PutUint16(b[idx:], sink16)
}

func store_be64(b []byte) {
	// amd64:`BSWAPQ`,-`SHR.`
	// arm64:`MOVD`,`REV`,-`MOV[WBH]`,-`REVW`,-`REV16W`
	binary.BigEndian.PutUint64(b, sink64)
}

func store_be64_idx(b []byte, idx int) {
	// amd64:`BSWAPQ`,-`SHR.`
	// arm64:`REV`,`MOVD\sR[0-9]+,\s\(R[0-9]+\)\(R[0-9]+\)`,-`MOV[BHW]`,-`REV16W`,-`REVW`
	binary.BigEndian.PutUint64(b[idx:], sink64)
}

func store_be32(b []byte) {
	// amd64:`BSWAPL`,-`SHR.`
	// arm64:`MOVW`,`REVW`,-`MOV[BH]`,-`REV16W`
	binary.BigEndian.PutUint32(b, sink32)
}

func store_be32_idx(b []byte, idx int) {
	// amd64:`BSWAPL`,-`SHR.`
	// arm64:`REVW`,`MOVW\sR[0-9]+,\s\(R[0-9]+\)\(R[0-9]+\)`,-`MOV[BH]`,-`REV16W`
	binary.BigEndian.PutUint32(b[idx:], sink32)
}

func store_be16(b []byte) {
	// amd64:`ROLW\s\$8`,-`SHR.`
	// arm64:`MOVH`,`REV16W`,-`MOVB`
	binary.BigEndian.PutUint16(b, sink16)
}

func store_be16_idx(b []byte, idx int) {
	// amd64:`ROLW\s\$8`,-`SHR.`
	// arm64:`MOVH\sR[0-9]+,\s\(R[0-9]+\)\(R[0-9]+\)`,`REV16W`,-`MOVB`
	binary.BigEndian.PutUint16(b[idx:], sink16)
}

// ------------- //
//    Zeroing    //
// ------------- //

// Check that zero stores are combined into larger stores

func zero_byte_2(b1, b2 []byte) {
	// bounds checks to guarantee safety of writes below
	_, _ = b1[1], b2[1]
	b1[0], b1[1] = 0, 0 // arm64:"MOVH\tZR",-"MOVB"
	b2[1], b2[0] = 0, 0 // arm64:"MOVH\tZR",-"MOVB"
}

func zero_byte_4(b1, b2 []byte) {
	_, _ = b1[3], b2[3]
	b1[0], b1[1], b1[2], b1[3] = 0, 0, 0, 0 // arm64:"MOVW\tZR",-"MOVB",-"MOVH"
	b2[2], b2[3], b2[1], b2[0] = 0, 0, 0, 0 // arm64:"MOVW\tZR",-"MOVB",-"MOVH"
}

func zero_byte_8(b []byte) {
	_ = b[7]
	b[0], b[1], b[2], b[3] = 0, 0, 0, 0
	b[4], b[5], b[6], b[7] = 0, 0, 0, 0 // arm64:"MOVD\tZR",-"MOVB",-"MOVH",-"MOVW"
}

func zero_byte_16(b []byte) {
	_ = b[15]
	b[0], b[1], b[2], b[3] = 0, 0, 0, 0
	b[4], b[5], b[6], b[7] = 0, 0, 0, 0
	b[8], b[9], b[10], b[11] = 0, 0, 0, 0
	b[12], b[13], b[14], b[15] = 0, 0, 0, 0 // arm64:"STP",-"MOVB",-"MOVH",-"MOVW"
}

func zero_byte_30(a *[30]byte) {
	*a = [30]byte{} // arm64:"STP",-"MOVB",-"MOVH",-"MOVW"
}

func zero_byte_39(a *[39]byte) {
	*a = [39]byte{} // arm64:"MOVD",-"MOVB",-"MOVH",-"MOVW"
}

func zero_uint16_2(h1, h2 []uint16) {
	_, _ = h1[1], h2[1]
	h1[0], h1[1] = 0, 0 // arm64:"MOVW\tZR",-"MOVB",-"MOVH"
	h2[1], h2[0] = 0, 0 // arm64:"MOVW\tZR",-"MOVB",-"MOVH"
}

func zero_uint16_4(h1, h2 []uint16) {
	_, _ = h1[3], h2[3]
	h1[0], h1[1], h1[2], h1[3] = 0, 0, 0, 0 // arm64:"MOVD\tZR",-"MOVB",-"MOVH",-"MOVW"
	h2[2], h2[3], h2[1], h2[0] = 0, 0, 0, 0 // arm64:"MOVD\tZR",-"MOVB",-"MOVH",-"MOVW"
}

func zero_uint16_8(h []uint16) {
	_ = h[7]
	h[0], h[1], h[2], h[3] = 0, 0, 0, 0
	h[4], h[5], h[6], h[7] = 0, 0, 0, 0 // arm64:"STP",-"MOVB",-"MOVH"
}

func zero_uint32_2(w1, w2 []uint32) {
	_, _ = w1[1], w2[1]
	w1[0], w1[1] = 0, 0 // arm64:"MOVD\tZR",-"MOVB",-"MOVH",-"MOVW"
	w2[1], w2[0] = 0, 0 // arm64:"MOVD\tZR",-"MOVB",-"MOVH",-"MOVW"
}

func zero_uint32_4(w1, w2 []uint32) {
	_, _ = w1[3], w2[3]
	w1[0], w1[1], w1[2], w1[3] = 0, 0, 0, 0 // arm64:"STP",-"MOVB",-"MOVH"
	w2[2], w2[3], w2[1], w2[0] = 0, 0, 0, 0 // arm64:"STP",-"MOVB",-"MOVH"
}

func zero_uint64_2(d1, d2 []uint64) {
	_, _ = d1[1], d2[1]
	d1[0], d1[1] = 0, 0 // arm64:"STP",-"MOVB",-"MOVH"
	d2[1], d2[0] = 0, 0 // arm64:"STP",-"MOVB",-"MOVH"
}
