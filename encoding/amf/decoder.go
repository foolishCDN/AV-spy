package amf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

func NewDecoder(version int) *Decoder {
	decoder := &Decoder{
		refObjects: make([]interface{}, 0),
	}
	if version == Version3 {
		decoder.refStrings = make([]string, 0)
		decoder.refTraits = make([]TraitAMF3, 0)
	}
	decoder.Version = version
	return decoder
}

type Decoder struct {
	Version int

	u8         [1]byte
	u16        [2]byte
	u32        [4]byte
	u64        [8]byte
	refObjects []interface{}

	// only amf3
	refStrings []string
	refTraits  []TraitAMF3
}

func (decoder *Decoder) Decode(r io.Reader) (interface{}, error) {
	if decoder.Version == Version3 {
		return decoder.decodeAMF3(r)
	}
	return decoder.decodeAMF0(r)
}

func (decoder *Decoder) DecodeBatch(r io.Reader) (res []interface{}, err error) {
	var v interface{}
	for {
		if decoder.Version == Version3 {
			v, err = decoder.decodeAMF3(r)
		} else {
			v, err = decoder.decodeAMF0(r)
		}
		if err != nil {
			break
		}
		res = append(res, v)
	}
	return res, err
}

func (decoder *Decoder) decodeAMF0(r io.Reader) (interface{}, error) {
	if _, err := io.ReadFull(r, decoder.u8[:]); err != nil {
		return nil, err
	}
	switch decoder.u8[0] {
	case NumberMarker:
		return decoder.readNumberOrAMF3Double(r)
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
		return decoder.decodeAMF3(r)
	default:
		return nil, fmt.Errorf("amf0 decoder: unsupported type %x", decoder.u8[0])
	}
}

// The data following a Number type marker is always an 8 byte IEEE-754 double precision floating point value
// in network byte order (sign bit in low memory)
func (decoder *Decoder) readNumberOrAMF3Double(r io.Reader) (float64, error) {
	if _, err := io.ReadFull(r, decoder.u64[:]); err != nil {
		return 0, err
	}
	u64 := binary.BigEndian.Uint64(decoder.u64[:])
	number := math.Float64frombits(u64)
	return number, nil
}

// A Boolean type marker is followed by an unsigned byte;
// a zero byte value denotes false while a non-zero byte value (typically 1) denotes true.
func (decoder *Decoder) readBoolean(r io.Reader) (bool, error) {
	if _, err := io.ReadFull(r, decoder.u8[:]); err != nil {
		return false, err
	}
	return decoder.u8[0] != 0, nil
}

// The AMF 0 Object type is used to encoded anonymous ActionScript objects.
// Any typed object that does not have a registered class should be treated as an anonymous ActionScript object.
// If the same object instance appears in an object graph it should be sent by reference using an AMF 0.
// Use the reference type to reduce redundant information from being serialized and infinite loops from cyclical references.
//
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
			if _, err := io.ReadFull(r, decoder.u8[:]); err != nil {
				return nil, err
			}
			if decoder.u8[0] == ObjectEndMarker {
				break
			} else {
				return nil, errors.New("should have ObjectEndMarker")
			}
		}
		value, err := decoder.decodeAMF0(r)
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
	if _, err := io.ReadFull(r, decoder.u16[:]); err != nil {
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
func (decoder *Decoder) readECMAArray(r io.Reader) (ECMAArray, error) {
	if _, err := io.ReadFull(r, decoder.u32[:]); err != nil {
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
	if _, err := io.ReadFull(r, decoder.u32[:]); err != nil {
		return nil, err
	}
	arrayCount := binary.BigEndian.Uint32(decoder.u32[:])
	array := make([]interface{}, arrayCount)
	var err error
	for i := uint32(0); i < arrayCount; i++ {
		array[i], err = decoder.decodeAMF0(r)
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
	if _, err := io.ReadFull(r, decoder.u64[:]); err != nil {
		return 0, err
	}
	u64 := binary.BigEndian.Uint64(decoder.u64[:])
	date := math.Float64frombits(u64)
	if _, err := io.ReadFull(r, decoder.u16[:]); err != nil { // time zone, ignore
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
	if _, err := io.ReadFull(r, decoder.u16[:]); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint16(decoder.u16[:])
	if length == 0 {
		return "", nil
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}

func (decoder *Decoder) readUTF8Long(r io.Reader) (string, error) {
	if _, err := io.ReadFull(r, decoder.u32[:]); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint32(decoder.u32[:])
	if length == 0 {
		return "", nil
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}

func (decoder *Decoder) decodeAMF3(r io.Reader) (interface{}, error) {
	if _, err := io.ReadFull(r, decoder.u8[:]); err != nil {
		return nil, err
	}
	switch decoder.u8[0] {
	case UndefinedMarkerAMF3:
		return UndefinedType{}, nil
	case NullMarkerAMF3:
		return NullType{}, nil
	case FalseMarkerAMF3:
		return FalseTypeAMF3{}, nil
	case TrueMarkerAMF3:
		return TrueTypeAMF3{}, nil
	case IntegerMarkerAMF3:
		n, err := DecodeUint29(r)
		if err != nil {
			return nil, err
		}
		return n, nil
	case DoubleMarkerAMF3:
		return decoder.readNumberOrAMF3Double(r)
	case StringMarkerAMF3:
		return decoder.readStringAMF3(r)
	case XMLDocMarkerAMF3:
		return decoder.readXMLDocAMF3(r)
	case DateMarkerAMF3:
		return decoder.readDateAMF3(r)
	case ArrayMarkerAMF3:
		return decoder.readArrayAMF3(r)
	case ObjectMarkerAMF3:
		return decoder.readObjectAMF3(r)
	case XMLMarkerAMF3:
		return decoder.readXMLAMF3(r)
	case ByteArrayMarkerAMF3:
		return decoder.readByteArrayAMF3(r)
	default:
		return nil, fmt.Errorf("amf3 decoder: unsupported type %x", decoder.u8[0])
	}
}

func (decoder *Decoder) readStringAMF3(r io.Reader) (string, error) {
	isRef, index, err := readRefInt(r)
	if err != nil {
		return "", err
	}
	if isRef {
		str, err := decoder.getRefString(index)
		if err != nil {
			return "", err
		}
		return str, nil
	}
	b := make([]byte, index)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	str := string(b)
	if str != "" {
		decoder.refStrings = append(decoder.refStrings, str)
	}
	return str, nil
}

func (decoder *Decoder) readXMLDocAMF3(r io.Reader) (XMLDocumentType, error) {
	isRef, index, err := readRefInt(r)
	if err != nil {
		return "", err
	}
	if isRef {
		object, err := decoder.getRefObject(index)
		if err != nil {
			return "", err
		}
		if xmlDoc, ok := object.(XMLDocumentType); ok {
			return xmlDoc, nil
		}
		return "", errors.New("readXMLDocAMF3: wrong type")
	}
	b := make([]byte, index)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	xmlDoc := XMLDocumentType(b)
	decoder.refObjects = append(decoder.refObjects, xmlDoc)
	return xmlDoc, nil
}

func (decoder *Decoder) readDateAMF3(r io.Reader) (DateType, error) {
	isRef, index, err := readRefInt(r)
	if err != nil {
		return 0, err
	}
	if isRef {
		object, err := decoder.getRefObject(index)
		if err != nil {
			return 0, err
		}
		if date, ok := object.(DateType); ok {
			return date, nil
		}
		return 0, errors.New("readDateAMF3: wrong type")
	}
	d, err := decoder.readNumberOrAMF3Double(r)
	if err != nil {
		return 0, err
	}
	date := DateType(d)
	decoder.refObjects = append(decoder.refObjects, date)
	return date, nil
}

func (decoder *Decoder) readArrayAMF3(r io.Reader) (*ArrayTypeAMF3, error) {
	isRef, index, err := readRefInt(r)
	if err != nil {
		return nil, err
	}
	if isRef {
		object, err := decoder.getRefObject(index)
		if err != nil {
			return nil, err
		}
		if array, ok := object.(*ArrayTypeAMF3); ok {
			return array, nil
		}
		return nil, errors.New("readArrayAMF3: wrong type")
	}
	array := &ArrayTypeAMF3{
		Associative: make(map[string]interface{}, 0),
		Dense:       make([]interface{}, index),
	}
	for {
		str, err := decoder.readStringAMF3(r)
		if err != nil {
			return nil, err
		}
		if str == "" {
			break
		}
		array.Associative[str], err = decoder.decodeAMF3(r)
		if err != nil {
			return nil, err
		}
	}
	for i := uint32(0); i < index; i++ {
		array.Dense[i], err = decoder.decodeAMF3(r)
		if err != nil {
			return nil, err
		}
	}
	decoder.refObjects = append(decoder.refObjects, array)
	return array, nil
}

func (decoder *Decoder) readObjectAMF3(r io.Reader) (*ObjectTypeAMF3, error) {
	isRef, index, err := readRefInt(r)
	if err != nil {
		return nil, err
	}
	if isRef {
		object, err := decoder.getRefObject(index)
		if err != nil {
			return nil, err
		}
		if array, ok := object.(*ObjectTypeAMF3); ok {
			return array, nil
		}
		return nil, errors.New("readObjectAMF3: wrong type")
	}
	var trait *TraitAMF3
	switch {
	case index&0x01 == 0:
		trait, err = decoder.getRefTrait(index >> 1)
		if err != nil {
			return nil, err
		}
	case index&0x02 == 0:
		trait := TraitAMF3{
			Attributes: make([]string, index>>3),
		}
		trait.IsDynamic = index&0x04 == 1
		trait.ClassName, err = decoder.readStringAMF3(r)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(trait.Attributes); i++ {
			trait.Attributes[i], err = decoder.readStringAMF3(r)
			if err != nil {
				return nil, err
			}
		}
		decoder.refTraits = append(decoder.refTraits, trait)
	default:
		return nil, fmt.Errorf("readObjectAMF3: index %b not supported", index)
	}
	obj := &ObjectTypeAMF3{
		Trait:  trait,
		Static: make([]interface{}, len(trait.Attributes)),
	}
	for i := 0; i < len(trait.Attributes); i++ {
		obj.Static[i], err = decoder.decodeAMF3(r)
		if err != nil {
			return nil, err
		}
	}
	obj.Dynamic = make(map[string]interface{})
	if trait.IsDynamic {
		for {
			name, err := decoder.readStringAMF3(r)
			if err != nil {
				return nil, err
			}
			if name == "" {
				break
			}
			obj.Dynamic[name], err = decoder.decodeAMF3(r)
			if err != nil {
				return nil, err
			}
		}
	}
	decoder.refObjects = append(decoder.refObjects, obj)
	return obj, nil
}

func (decoder *Decoder) readXMLAMF3(r io.Reader) (XMLTypeAMF3, error) {
	isRef, index, err := readRefInt(r)
	if err != nil {
		return "", err
	}
	if isRef {
		object, err := decoder.getRefObject(index)
		if err != nil {
			return "", err
		}
		if xml, ok := object.(XMLTypeAMF3); ok {
			return xml, nil
		}
		return "", errors.New("readXMLAMF3: wrong type")
	}
	b := make([]byte, index)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	xml := XMLTypeAMF3(b)
	decoder.refObjects = append(decoder.refObjects, xml)
	return xml, nil
}

func (decoder *Decoder) readByteArrayAMF3(r io.Reader) ([]byte, error) {
	isRef, index, err := readRefInt(r)
	if err != nil {
		return nil, err
	}
	if isRef {
		object, err := decoder.getRefObject(index)
		if err != nil {
			return nil, err
		}
		if xml, ok := object.([]byte); ok {
			return xml, nil
		}
		return nil, errors.New("readByteArrayAMF3: wrong type")
	}
	b := make([]byte, index)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	decoder.refObjects = append(decoder.refObjects, b)
	return b, nil
}

func (decoder *Decoder) getRefString(index uint32) (string, error) {
	if int(index) >= len(decoder.refStrings) {
		return "", fmt.Errorf("refStrings: index %d out of bound", index)
	}
	return decoder.refStrings[index], nil
}

func (decoder *Decoder) getRefObject(index uint32) (interface{}, error) {
	if int(index) >= len(decoder.refObjects) {
		return nil, fmt.Errorf("refObjects: index %d out of bound", index)
	}
	return decoder.refObjects[index], nil
}

func (decoder *Decoder) getRefTrait(index uint32) (*TraitAMF3, error) {
	if int(index) >= len(decoder.refTraits) {
		return nil, fmt.Errorf("refTraits: index %d out of bound", index)
	}
	return &decoder.refTraits[index], nil
}

func readRefInt(r io.Reader) (isRef bool, index uint32, err error) {
	u29, err := DecodeUint29(r)
	if err != nil {
		return false, 0, err
	}
	isRef = u29&0x01 == 0
	index = u29 >> 1
	return isRef, index, nil
}
