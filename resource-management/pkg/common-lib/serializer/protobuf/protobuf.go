/*
Copyright 2015 The Kubernetes Authors.
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

package protobuf

import (
	"fmt"
	"io"
	"reflect"

	"github.com/gogo/protobuf/proto"
	"k8s.io/klog/v2"

	"global-resource-service/resource-management/pkg/common-lib/serializer"
)

type errNotMarshalable struct {
	t reflect.Type
}

func (e errNotMarshalable) Error() string {
	return fmt.Sprintf("object %v does not implement the protobuf marshalling interface and cannot be encoded to a protobuf message", e.t)
}

// IsNotMarshalable checks the type of error, returns a boolean true if error is not nil and not marshalable false otherwise
func IsNotMarshalable(err error) bool {
	_, ok := err.(errNotMarshalable)
	return err != nil && ok
}

// NewSerializer creates a Protobuf serializer that handles encoding versioned objects into the proper wire form. If a typer
// is passed, the encoded object will have group, version, and kind fields set. If typer is nil, the objects will be written
// as-is (any type info passed with the object will be used).
func NewSerializer(typer serializer.ObjectTyper) *Serializer {
	return &Serializer{
		typer: typer,
	}
}

// Serializer handles encoding versioned objects into the proper wire form
type Serializer struct {
	prefix []byte
	typer  serializer.ObjectTyper
}

var _ serializer.Serializer = &Serializer{}

const serializerIdentifier serializer.Identifier = "protobuf"

// Decode attempts to convert the provided data into a protobuf message, extract the stored schema kind, apply the provided default
// gvk, and then load that data into an object matching the desired schema kind or the provided into. If into is *serializer.Unknown,
// the raw data will be extracted and no decoding will be performed. If into is not registered with the typer, then the object will
// be straight decoded using normal protobuf unmarshalling (the MarshalTo interface). If into is provided and the original data is
// not fully qualified with kind/version/group, the type of the into will be used to alter the returned gvk. On success or most
// errors, the method will return the calculated schema kind.
func (s *Serializer) Decode(data []byte, into interface{}) (interface{}, error) {

	err := proto.Unmarshal(data, into.(proto.Message))

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data")
	}

	return into, nil
}

// Encode serializes the provided object to the given writer.
func (s *Serializer) Encode(obj interface{}, w io.Writer) error {
	b, err := proto.Marshal(obj.(proto.Message))

	if err != nil {
		klog.Errorf("failed to marshal object, error %v", err)
		return err
	}

	_, err = w.Write(b)
	if err != nil {
		klog.Errorf("failed to write response, error %v", err)
		return err
	}

	return nil
}

func (s *Serializer) Marshal(obj interface{}) (b []byte, err error) {
	b, err = proto.Marshal(obj.(proto.Message))

	if err != nil {
		klog.Errorf("failed to marshal object, error %v", err)
		return nil, err
	}

	return b, nil
}

// Identifier implements serializer.Encoder interface.
func (s *Serializer) Identifier() serializer.Identifier {
	return serializerIdentifier
}
