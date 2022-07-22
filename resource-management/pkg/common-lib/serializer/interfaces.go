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

//// MemoryAllocator is responsible for allocating memory.
//// By encapsulating memory allocation into its own interface, we can reuse the memory
//// across many operations in places we know it can significantly improve the performance.
//type MemoryAllocator interface {
//	// Allocate reserves memory for n bytes.
//	// Note that implementations of this method are not required to zero the returned array.
//	// It is the caller's responsibility to clean the memory if needed.
//	Allocate(n uint64) []byte
//}
//
//// EncoderWithAllocator  serializes objects in a way that allows callers to manage any additional memory allocations.
//type EncoderWithAllocator interface {
//	Encoder
//	// EncodeWithAllocator writes an object to a stream as Encode does.
//	// In addition, it allows for providing a memory allocator for efficient memory usage during object serialization
//	EncodeWithAllocator(obj interface{}, w io.Writer, memAlloc MemoryAllocator) error
//}

// Decoder attempts to load an object from data.
type Decoder interface {
	Decode([]byte, interface{}) (interface{}, error)
}

// Serializer is the core interface for transforming objects into a serialized format and back.
// Implementations may choose to perform conversion of the object, but no assumptions should be made.
type Serializer interface {
	Encoder
	Decoder
}

// Codec is a Serializer that deals with the details of versioning objects. It offers the same
// interface as Serializer, so this is a marker to consumers that care about the version of the objects
// they receive.
// type Codec Serializer

//// ParameterCodec defines methods for serializing and deserializing API objects to url.Values and
//// performing any necessary conversion. Unlike the normal Codec, query parameters are not self describing
//// and the desired version must be specified.
//type ParameterCodec interface {
//	// DecodeParameters takes the given url.Values in the specified group version and decodes them
//	// into the provided object, or returns an error.
//	DecodeParameters(parameters url.Values, into interface{}) error
//	// EncodeParameters encodes the provided object as query parameters or returns an error.
//	EncodeParameters(obj interface{}) (url.Values, error)
//}

// Framer is a factory for creating readers and writers that obey a particular framing pattern.
type Framer interface {
	NewFrameReader(r io.ReadCloser) io.ReadCloser
	NewFrameWriter(w io.Writer) io.Writer
}

//// SerializerInfo contains information about a specific serialization format
//type SerializerInfo struct {
//	// MediaType is the value that represents this serializer over the wire.
//	MediaType string
//	// MediaTypeType is the first part of the MediaType ("application" in "application/json").
//	MediaTypeType string
//	// MediaTypeSubType is the second part of the MediaType ("json" in "application/json").
//	MediaTypeSubType string
//	// EncodesAsText indicates this serializer can be encoded to UTF-8 safely.
//	EncodesAsText bool
//	// Serializer is the individual object serializer for this media type.
//	Serializer Serializer
//	// PrettySerializer, if set, can serialize this object in a form biased towards
//	// readability.
//	PrettySerializer Serializer
//	// StrictSerializer, if set, deserializes this object strictly,
//	// erring on unknown fields.
//	StrictSerializer Serializer
//	// StreamSerializer, if set, describes the streaming serialization format
//	// for this media type.
//	StreamSerializer *StreamSerializerInfo
//}
//
// StreamSerializerInfo contains information about a specific stream serialization format
//type StreamSerializerInfo struct {
//	// EncodesAsText indicates this serializer can be encoded to UTF-8 safely.
//	EncodesAsText bool
//	// Serializer is the top level object serializer for this type when streaming
//	Serializer
//	// Framer is the factory for retrieving streams that separate objects on the wire
//	Framer
//}

//// NegotiatedSerializer is an interface used for obtaining encoders, decoders, and serializers
//// for multiple supported media types. This would commonly be accepted by a server component
//// that performs HTTP content negotiation to accept multiple formats.
//type NegotiatedSerializer interface {
//	// SupportedMediaTypes is the media types supported for reading and writing single objects.
//	SupportedMediaTypes() []SerializerInfo
//
//	// EncoderForVersion returns an encoder that ensures objects being written to the provided
//	// serializer are in the provided group version.
//	EncoderForVersion(serializer Encoder) Encoder
//	// DecoderToVersion returns a decoder that ensures objects being read by the provided
//	// serializer are in the provided group version by default.
//	DecoderToVersion(serializer Decoder) Decoder
//}

//// ClientNegotiator handles turning an HTTP content type into the appropriate encoder.
//// Use NewClientNegotiator or NewVersionedClientNegotiator to create this interface from
//// a NegotiatedSerializer.
//type ClientNegotiator interface {
//	// Encoder returns the appropriate encoder for the provided contentType (e.g. application/json)
//	// and any optional mediaType parameters (e.g. pretty=1), or an error. If no serializer is found
//	// a NegotiateError will be returned. The current client implementations consider params to be
//	// optional modifiers to the contentType and will ignore unrecognized parameters.
//	Encoder(contentType string, params map[string]string) (Encoder, error)
//	// Decoder returns the appropriate decoder for the provided contentType (e.g. application/json)
//	// and any optional mediaType parameters (e.g. pretty=1), or an error. If no serializer is found
//	// a NegotiateError will be returned. The current client implementations consider params to be
//	// optional modifiers to the contentType and will ignore unrecognized parameters.
//	Decoder(contentType string, params map[string]string) (Decoder, error)
//	// StreamDecoder returns the appropriate stream decoder for the provided contentType (e.g.
//	// application/json) and any optional mediaType parameters (e.g. pretty=1), or an error. If no
//	// serializer is found a NegotiateError will be returned. The Serializer and Framer will always
//	// be returned if a Decoder is returned. The current client implementations consider params to be
//	// optional modifiers to the contentType and will ignore unrecognized parameters.
//	StreamDecoder(contentType string, params map[string]string) (Decoder, Serializer, Framer, error)
//}

//// StorageSerializer is an interface used for obtaining encoders, decoders, and serializers
//// that can read and write data at rest. This would commonly be used by client tools that must
//// read files, or server side storage interfaces that persist restful objects.
//type StorageSerializer interface {
//	// SupportedMediaTypes are the media types supported for reading and writing objects.
//	SupportedMediaTypes() []SerializerInfo
//
//	// UniversalDeserializer returns a Serializer that can read objects in multiple supported formats
//	// by introspecting the data at rest.
//	UniversalDeserializer() Decoder
//
//	// EncoderForVersion returns an encoder that ensures objects being written to the provided
//	// serializer are in the provided group version.
//	EncoderForVersion(serializer Encoder) Encoder
//	// DecoderForVersion returns a decoder that ensures objects being read by the provided
//	// serializer are in the provided group version by default.
//	DecoderToVersion(serializer Decoder) Decoder
//}

//// for POC, simply without versioning support of the node object
//type ObjectCreater interface {
//	New() (out interface{}, err error)
//}
