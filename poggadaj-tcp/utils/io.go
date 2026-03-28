// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

// Derived from https://github.com/hexis-revival/hexagon/blob/main/common/io.go
// licensed under the MIT license. Adjusted for the GG protocol.

package utils

import (
	"bytes"
	"encoding/binary"
	"math"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type IOStream struct {
	data     []byte
	position int
	endian   binary.ByteOrder
	encoding *charmap.Charmap
}

func NewIOStream(data []byte, endian binary.ByteOrder, encoding *charmap.Charmap) *IOStream {
	return &IOStream{
		data:     data,
		position: 0,
		endian:   endian,
		encoding: encoding,
	}
}

func (stream *IOStream) Push(data []byte) {
	stream.data = append(stream.data, data...)
}

func (stream *IOStream) Get() []byte {
	return stream.data
}

func (stream *IOStream) Len() int {
	return len(stream.data)
}

func (stream *IOStream) Available() int {
	return stream.Len() - stream.position
}

func (stream *IOStream) Tell() int {
	return stream.position
}

func (stream *IOStream) Seek(position int) {
	stream.position = position
}

func (stream *IOStream) Skip(offset int) {
	stream.position += offset
}

func (stream *IOStream) Eof() bool {
	return stream.position >= stream.Len()
}

func (stream *IOStream) Read(size int) []byte {
	if stream.Eof() {
		return []byte{}
	}

	if stream.Available() < size {
		size = stream.Available()
	}

	data := stream.data[stream.position : stream.position+size]
	stream.position += size

	return data
}

func (stream *IOStream) ReadAll() []byte {
	return stream.Read(stream.Available())
}

func (stream *IOStream) ReadU8() uint8 {
	return stream.Read(1)[0]
}

func (stream *IOStream) ReadU16() uint16 {
	return stream.endian.Uint16(stream.Read(2))
}

func (stream *IOStream) ReadU32() uint32 {
	return stream.endian.Uint32(stream.Read(4))
}

func (stream *IOStream) ReadU64() uint64 {
	return stream.endian.Uint64(stream.Read(8))
}

func (stream *IOStream) ReadI8() int8 {
	return int8(stream.Read(1)[0])
}

func (stream *IOStream) ReadI16() int16 {
	return int16(stream.ReadU16())
}

func (stream *IOStream) ReadI32() int32 {
	return int32(stream.ReadU32())
}

func (stream *IOStream) ReadI64() int64 {
	return int64(stream.ReadU64())
}

func (stream *IOStream) ReadF32() float32 {
	bits := stream.ReadU32()
	return math.Float32frombits(bits)
}

func (stream *IOStream) ReadF64() float64 {
	bits := stream.ReadU64()
	return math.Float64frombits(bits)
}

// ReadString reads a GG string of specified size and with IOStream's encoding.
// If length is -1, then the string gets read until EOF.
func (stream *IOStream) ReadString(size int) string {
	if size == -1 {
		size = stream.Available()
	}

	str, _ := stream.encoding.NewDecoder().Bytes(stream.Read(size))

	return string(str)
}

func (stream *IOStream) Write(data []byte) {
	stream.Push(data)
}

func (stream *IOStream) WriteU8(value uint8) {
	stream.Write([]byte{value})
}

func (stream *IOStream) WriteU16(value uint16) {
	data := make([]byte, 2)
	stream.endian.PutUint16(data, value)
	stream.Write(data)
}

func (stream *IOStream) WriteU32(value uint32) {
	data := make([]byte, 4)
	stream.endian.PutUint32(data, value)
	stream.Write(data)
}

func (stream *IOStream) WriteU64(value uint64) {
	data := make([]byte, 8)
	stream.endian.PutUint64(data, value)
	stream.Write(data)
}

func (stream *IOStream) WriteI8(value int8) {
	stream.WriteU8(uint8(value))
}

func (stream *IOStream) WriteI16(value int16) {
	stream.WriteU16(uint16(value))
}

func (stream *IOStream) WriteI32(value int32) {
	stream.WriteU32(uint32(value))
}

func (stream *IOStream) WriteI64(value int64) {
	stream.WriteU64(uint64(value))
}

func (stream *IOStream) WriteF32(value float32) {
	bits := math.Float32bits(value)
	stream.WriteU32(bits)
}

func (stream *IOStream) WriteF64(value float64) {
	bits := math.Float64bits(value)
	stream.WriteU64(bits)
}

// SerializeString serializes a string using IOStream's encoding.
// It does NOT write the string to the IOStream.
// nullTerminate indicates whether the string should be null-terminated.
func (stream *IOStream) SerializeString(value string, nullTerminate bool) []byte {
	var buf bytes.Buffer

	encoder := transform.NewWriter(&buf, stream.encoding.NewEncoder())
	encoder.Write([]byte(value))

	if nullTerminate {
		encoder.Write([]byte("\x00"))
	}

	return buf.Bytes()
}
