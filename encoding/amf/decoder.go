package amf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

func NewDecoder() *Decoder {
	return &Decoder{
		refObjects: make([]interface{}, 0),
	}
}

type Decoder struct {
	u8         [1]byte
	u16        [2]byte
	u32        [4]byte
	u64        [8]byte
	refObjects []interface{}
}

func (decoder *Decoder) Decode(r io.Reader) (interface{}, error) {
	return decoder.decode(r)
}

func (decoder *Decoder) DecodeBatch(r io.Reader) (res []interface{}, err error) {
	var v interface{}
	for {
		v, err = decoder.decode(r)
		if err != nil {
			break
		}
		res = append(res, v)
	}
	return res, err
}

func (decoder *Decoder) decode(r io.Reader) (interface{}, error) {
	if _, err := r.Read(decoder.u8[:]); err != nil {
		return nil, err
	}
	marker := decoder.u8[0]
	switch marker {
	case NumberMarker:
		return decoder.readNumber(r)
	case BooleanMarker:
		return decoder.readBoolean(r)
	case StringMarker:
		return decoder.readUTF8(r)
	case ObjectMarker:
		obj, err := decoder.readObject(r)
		if err != nil {
			return nil, err
		}
		decoder.refObjects = append(decoder.refObjects, obj)
		return obj, nil
	case MovieClipMarker:
		// This type is not supported and is reserved for future use.
		return nil, errors.New("movieClip type is not supported")
	case NullMarker:
		return NullType{}, nil
	case UndefinedMarker:
		return UndefinedType{}, nil
	case ReferenceMarker:
		return decoder.readReference(r)
	case ECMAArrayMarker:
		return decoder.readECMAArray(r)
	case StrictArrayMarker:
		return decoder.readStrictArray(r)
	case DateMarker:
		return decoder.readDate(r)
	case LongStringMarker:
		return decoder.readUTF8Long(r)
	case UnSupportedMarker:
		return UnsupportedType{}, nil
	case RecordSetMarker:
		// This type is not supported and is reserved for future use.
		return nil, errors.New("recordSet type is not supported")
	case XMLDocumentMarker:
		// The XML document type is always encoded as a long UTF-8 string.
		str, err := decoder.readUTF8Long(r)
		if err != nil {
			return nil, err
		}
		return XMLDocumentType(str), nil
	case TypedObjectMarker:
		return decoder.readTypedObject(r)
	case AVMPlusObjectMarker:
		// TODO: support amf3
		return nil, errors.New("amf0 is not supported")
	default:
		return nil, errors.New("should not reach")
	}
}

// The data following a Number type marker is always an 8 byte IEEE-754 double precision floating point value
// in network byte order (sign bit in low memory)
func (decoder *Decoder) readNumber(r io.Reader) (float64, error) {
	if _, err := r.Read(decoder.u64[:]); err != nil {
		return 0, err
	}
	u64 := binary.BigEndian.Uint64(decoder.u64[:])
	number := math.Float64frombits(u64)
	return number, nil
}

// A Boolean type marker is followed by an unsigned byte;
// a zero byte value denotes false while a non-zero byte value (typically 1) denotes true.
func (decoder *Decoder) readBoolean(r io.Reader) (bool, error) {
	if _, err := r.Read(decoder.u8[:]); err != nil {
		return false, err
	}
	return decoder.u8[0] != 0, nil
}

// The AMF 0 Object type is used to encoded anonymous ActionScript objects.
// Any typed object that does not have a registered class should be treated as an anonymous ActionScript object.
// If the same object instance appears in an object graph it should be sent by reference using an AMF 0.
// Use the reference type to reduce redundant information from being serialized and infinite loops from cyclical references.
// object-property = (UTF-8 value-type) | (UTF-8-empty object-end-marker)
// anonymous-object-type = object-marker *(object-property)
func (decoder *Decoder) readObject(r io.Reader) (map[string]interface{}, error) {
	v := make(map[string]interface{})
	for {
		name, err := decoder.readUTF8(r)
		if err != nil {
			return nil, err
		}
		if name == "" {
			if _, err := r.Read(decoder.u8[:]); err != nil {
				return nil, err
			}
			if decoder.u8[0] == ObjectEndMarker {
				break
			} else {
				return nil, errors.New("should have ObjectEndMarker")
			}
		}
		value, err := decoder.decode(r)
		if err != nil {
			return nil, err
		}
		if _, ok := v[name]; ok {
			return nil, fmt.Errorf("object property %s exists", name)
		}
		v[name] = value
	}
	return v, nil
}

// AMF0 defines a complex object as an anonymous object, a typed object, an array or an ecma-array.
// If the exact same instance of a complex object appears more than once in an object graph
// then it must be sent by reference.
// The reference type uses an unsigned 16-bit integer to point to an index in a table of previously serialized objects.
// Indices start at 0.
//
// reference-type = reference-marker U16 ; index pointing to another complex type
//
// A 16-bit unsigned integer implies a theoretical maximum of 65,535 unique complex objects that can be sent by reference.
func (decoder *Decoder) readReference(r io.Reader) (interface{}, error) {
	if _, err := r.Read(decoder.u16[:]); err != nil {
		return nil, err
	}
	id := binary.BigEndian.Uint16(decoder.u16[:])
	if int(id) >= len(decoder.refObjects) {
		return nil, errors.New("reference error")
	}
	return decoder.refObjects[id], nil
}

// An ECMA Array or 'associative' Array is used when an ActionScript Array contains non-ordinal indices.
// This type is considered a complex type and thus reoccurring instances can be sent by reference.
// All indices, ordinal or otherwise, are treated as string 'keys' instead of integers.
// For the purposes of serialization this type is very similar to an anonymous Object.
//
// associative-count = U32
// ecma-array-type = associative-count *(object-property)
//
// A 32-bit associative-count implies a theoretical maximum of 4,294,967,295 associative array entries.
func (decoder *Decoder) readECMAArray(r io.Reader) (map[string]interface{}, error) {
	if _, err := r.Read(decoder.u32[:]); err != nil {
		return nil, err
	}
	associativeCount := binary.BigEndian.Uint32(decoder.u32[:])
	obj, err := decoder.readObject(r)
	if err != nil {
		return nil, err
	}
	if uint32(len(obj)) != associativeCount {
		return nil, errors.New("ECMAArray Count error")
	}
	decoder.refObjects = append(decoder.refObjects, obj)
	return obj, nil
}

// A strict Array contains only ordinal indices; however, in AMF 0 the indices can be dense or sparse.
// Undefined entries in the sparse regions between indices are serialized as undefined.
//
// array-count = U32
// strict-array-type = array-count *(value-type)
//
// A 32-bit array-count implies a theoretical maximum of 4,294,967,295 array entries.
func (decoder *Decoder) readStrictArray(r io.Reader) ([]interface{}, error) {
	if _, err := r.Read(decoder.u32[:]); err != nil {
		return nil, err
	}
	arrayCount := binary.BigEndian.Uint32(decoder.u32[:])
	array := make([]interface{}, arrayCount)
	var err error
	for i := uint32(0); i < arrayCount; i++ {
		array[i], err = decoder.decode(r)
		if err != nil {
			return nil, err
		}
	}
	decoder.refObjects = append(decoder.refObjects, array)
	return array, nil
}

// An ActionScript Date is serialized as
// the number of milliseconds elapsed since the epoch of midnight on 1st Jan 1970 in the UTC time zone.
// While the design of this type reserves room for time zone offset information,
// it should not be filled in, nor used,
// as it is unconventional to change time zones when serializing dates on a network.
// It is suggested that the time zone be queried independently as needed
//
// time-zone = S16 ; reserved, not supported should be set to 0x0000
// date-type = date-marker DOUBLE time-zone
func (decoder *Decoder) readDate(r io.Reader) (DateType, error) {
	if _, err := r.Read(decoder.u64[:]); err != nil {
		return 0, err
	}
	u64 := binary.BigEndian.Uint64(decoder.u64[:])
	date := math.Float64frombits(u64)
	if _, err := r.Read(decoder.u16[:]); err != nil { // time zone, ignore
		return 0, err
	}
	return DateType(date), nil
}

// If a strongly typed object has an alias registered for its class then the type name will also be serialized.
// Typed objects are considered complex types and reoccurring instances can be sent by reference.
//
// class-name = UTF-8
// object-type = object-marker class-name *(object-property)
func (decoder *Decoder) readTypedObject(r io.Reader) (*TypedObjectType, error) {
	className, err := decoder.readUTF8(r)
	if err != nil {
		return nil, err
	}
	obj, err := decoder.readObject(r)
	if err != nil {
		return nil, err
	}
	object := &TypedObjectType{
		ClassName: className,
		Object:    obj,
	}
	decoder.refObjects = append(decoder.refObjects, object)
	return object, nil
}

func (decoder *Decoder) readUTF8(r io.Reader) (string, error) {
	if _, err := r.Read(decoder.u16[:]); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint16(decoder.u16[:])
	if length == 0 {
		return "", nil
	}
	b := make([]byte, length)
	if _, err := r.Read(b); err != nil {
		return "", err
	}
	return string(b), nil
}

func (decoder *Decoder) readUTF8Long(r io.Reader) (string, error) {
	if _, err := r.Read(decoder.u32[:]); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint32(decoder.u32[:])
	if length == 0 {
		return "", nil
	}
	b := make([]byte, length)
	if _, err := r.Read(b); err != nil {
		return "", err
	}
	return string(b), nil
}
