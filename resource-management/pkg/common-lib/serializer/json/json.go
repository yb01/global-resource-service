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

package json

import (
	"encoding/json"
	"io"
	"strconv"

	"k8s.io/klog/v2"

	"global-resource-service/resource-management/pkg/common-lib/serializer"
)

func NewSerializer(typer serializer.ObjectTyper, pretty bool) *Serializer {
	return NewSerializerWithOptions(typer, SerializerOptions{false, false})
}

// NewSerializerWithOptions creates a JSON/YAML serializer that handles encoding versioned objects into the proper JSON/YAML
// form. If typer is not nil, the object has the group, version, and kind fields set. Options are copied into the Serializer
// and are immutable.
func NewSerializerWithOptions(typer serializer.ObjectTyper, options SerializerOptions) *Serializer {
	return &Serializer{
		typer:      typer,
		options:    options,
		identifier: identifier(options),
	}
}

// identifier computes Identifier of Encoder based on the given options.
func identifier(options SerializerOptions) serializer.Identifier {
	result := map[string]string{
		"name":   "json",
		"pretty": strconv.FormatBool(options.Pretty),
		"strict": strconv.FormatBool(options.Strict),
	}
	identifier, err := json.Marshal(result)
	if err != nil {
		klog.Fatalf("Failed marshaling identifier for json Serializer: %v", err)
	}
	return serializer.Identifier(identifier)
}

type SerializerOptions struct {
	// Pretty: configures a JSON enabled Serializer(`Yaml: false`) to produce human-readable output.
	Pretty bool

	// Strict: configures the Serializer to return strictDecodingError's when duplicate fields are present decoding JSON or YAML.
	// Note that enabling this option is not as performant as the non-strict variant, and should not be used in fast paths.
	Strict bool
}

// Serializer handles encoding versioned objects into the proper JSON form
type Serializer struct {
	options    SerializerOptions
	typer      serializer.ObjectTyper
	identifier serializer.Identifier
}

func (s *Serializer) Decode(data []byte, into interface{}) (interface{}, error) {

	err := s.Unmarshal(data, &into)

	return into, err
}

// Encode serializes the provided object to the given writer.
func (s *Serializer) Encode(obj interface{}, w io.Writer) error {
	return s.doEncode(obj, w)
}

func (s *Serializer) doEncode(obj interface{}, w io.Writer) error {
	if s.options.Pretty {
		data, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	encoder := json.NewEncoder(w)
	return encoder.Encode(obj)
}

// IsStrict indicates whether the serializer
// uses strict decoding or not
func (s *Serializer) IsStrict() bool {
	return s.options.Strict
}

func (s *Serializer) Unmarshal(data []byte, into *interface{}) (err error) {
	return json.Unmarshal(data, into)
}

func (s *Serializer) Marshal(obj interface{}) (b []byte, err error) {
	b, err = json.Marshal(obj)

	if err != nil {
		klog.Errorf("failed to marshal object, error %v", err)
		return nil, err
	}

	return b, nil
}

// Identifier implements serializer.Encoder interface.
func (s *Serializer) Identifier() serializer.Identifier {
	return s.identifier
}

//// Framer is the default JSON framing behavior, with newlines delimiting individual objects.
//var Framer = jsonFramer{}
//
//type jsonFramer struct{}
//
//// NewFrameWriter implements stream framing for this serializer
//func (jsonFramer) NewFrameWriter(w io.Writer) io.Writer {
//	// we can write JSON objects directly to the writer, because they are self-framing
//	return w
//}
//
//// NewFrameReader implements stream framing for this serializer
//func (jsonFramer) NewFrameReader(r io.ReadCloser) io.ReadCloser {
//	// we need to extract the JSON chunks of data to pass to Decode()
//	return framer.NewJSONFramedReader(r)
//}
