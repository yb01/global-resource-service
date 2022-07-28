/*
Copyright 2014 The Kubernetes Authors.
Copyright 2022 Authors of Arktos - file modified.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package serializer

import (
	"io"
)

// MemoryAllocator is responsible for allocating memory.
// By encapsulating memory allocation into its own interface, we can reuse the memory
// across many operations in places we know it can significantly improve the performance.
type MemoryAllocator interface {
	// Allocate reserves memory for n bytes.
	// Note that implementations of this method are not required to zero the returned array.
	// It is the caller's responsibility to clean the memory if needed.
	Allocate(n uint64) []byte
}

// SimpleAllocator a wrapper around make([]byte)
// conforms to the MemoryAllocator interface
type SimpleAllocator struct{}

var _ MemoryAllocator = &SimpleAllocator{}

func (sa *SimpleAllocator) Allocate(n uint64) []byte {
	return make([]byte, n, n)
}

type ObjectTyper string

// Identifier represents an identifier.
// Identitier of two different objects should be equal if and only if for every
// input the output they produce is exactly the same.
type Identifier string

// Encoder writes objects to a serialized form
type Encoder interface {
	Encode(interface{}, io.Writer) error
	Identifier() Identifier
}

// Decoder attempts to load an object from data.
type Decoder interface {
	Decode([]byte, interface{}) (interface{}, error)
}

// Serializer is the core interface for transforming objects into a serialized format and back.
// Implementations may choose to perform conversion of the object, but no assumptions should be made.
type Serializer interface {
	Encoder
	Decoder
	Marshal(interface{}) ([]byte, error)
}

// Framer is a factory for creating readers and writers that obey a particular framing pattern.
type Framer interface {
	NewFrameReader(r io.ReadCloser) io.ReadCloser
	NewFrameWriter(w io.Writer) io.Writer
}
