package amf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

func NewEncoder(version int) *Encoder {
	encoder := &Encoder{
		refObjects: make([]interface{}, 0),
	}
	if version == Version3 {
		encoder.refStrings = make([]string, 0)
		encoder.refTraits = make([]TraitAMF3, 0)
	}
	encoder.Version = version
	return encoder
}

type Encoder struct {
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

func (encoder *Encoder) Encode(w io.Writer, v interface{}) error {
	if encoder.Version == Version3 {
		return encoder.encodeAMF3(w, v)
	}
	return encoder.encodeAMF0(w, v)
}

func (encoder *Encoder) EncodeBatch(w io.Writer, vs ...interface{}) (err error) {
	for i := range vs {
		if encoder.Version == Version3 {
			err = encoder.encodeAMF3(w, vs[i])
		} else {
			err = encoder.encodeAMF0(w, vs[i])
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (encoder *Encoder) encodeAMF0(w io.Writer, v interface{}) error {
	switch t := v.(type) {
	case float64: // number type
		return encoder.EncodeNumberOrAMF3Double(w, t, NumberMarker)
	case bool: // boolean type
		return encoder.EncodeBoolean(w, t)
	case string: // string or long string type
		if len(t) <= math.MaxUint16 {
			return encoder.EncodeString(w, t)
		}
		return encoder.EncodeLongString(w, t)
	case map[string]interface{}: // object type
		return encoder.EncodeObject(w, t)
	case NullType:
		return encoder.EncodeMarker(w, NullMarker)
	case UndefinedType:
		return encoder.EncodeMarker(w, UndefinedMarker)
	case ECMAArray:
		return encoder.EncodeECMAArray(w, t)
	case []interface{}: // strict array type
		return encoder.EncodeStrictArray(w, t)
	case DateType:
		return encoder.EncodeDate(w, t)
	case UnsupportedType:
		return encoder.EncodeMarker(w, UnSupportedMarker)
	case XMLDocumentType:
		return encoder.EncodeXMLDocument(w, t)
	case *TypedObjectType:
		return encoder.EncodeTypedObject(w, t)
	default:
		return fmt.Errorf("type %T not supported", t)
	}
}

func (encoder *Encoder) EncodeMarker(w io.Writer, marker byte) error {
	encoder.u8[0] = marker
	_, err := w.Write(encoder.u8[:])
	return err
}

func (encoder *Encoder) EncodeNumberOrAMF3Double(w io.Writer, v float64, mark byte) error {
	if err := encoder.EncodeMarker(w, mark); err != nil {
		return err
	}
	number := math.Float64bits(v)
	binary.BigEndian.PutUint64(encoder.u64[:], number)
	_, err := w.Write(encoder.u64[:])
	return err
}

func (encoder *Encoder) EncodeBoolean(w io.Writer, b bool) (err error) {
	if err := encoder.EncodeMarker(w, BooleanMarker); err != nil {
		return err
	}
	if b {
		err = encoder.EncodeMarker(w, 1)
	} else {
		err = encoder.EncodeMarker(w, 0)
	}
	return err
}

func (encoder *Encoder) EncodeString(w io.Writer, str string) error {
	length := len(str)
	if length > math.MaxUint16 {
		return errors.New("string too long")
	}
	if err := encoder.EncodeMarker(w, StringMarker); err != nil {
		return err
	}
	return encoder.writeUTF8(w, str, uint16(length))
}

func (encoder *Encoder) EncodeObject(w io.Writer, m map[string]interface{}) error {
	encoder.refObjects = append(encoder.refObjects, m)
	return encoder.writeObject(w, m)
}

func (encoder *Encoder) EncodeECMAArray(w io.Writer, array ECMAArray) error {
	encoder.refObjects = append(encoder.refObjects, array)
	if err := encoder.EncodeMarker(w, ECMAArrayMarker); err != nil {
		return err
	}
	associativeCount := len(array)
	binary.BigEndian.PutUint32(encoder.u32[:], uint32(associativeCount))
	if _, err := w.Write(encoder.u32[:]); err != nil {
		return err
	}
	return encoder.writeObject(w, array)
}

func (encoder *Encoder) EncodeStrictArray(w io.Writer, array []interface{}) error {
	encoder.refObjects = append(encoder.refObjects, array)
	if err := encoder.EncodeMarker(w, StrictArrayMarker); err != nil {
		return err
	}
	arrayCount := len(array)
	binary.BigEndian.PutUint32(encoder.u32[:], uint32(arrayCount))
	if _, err := w.Write(encoder.u32[:]); err != nil {
		return err
	}
	for i := range array {
		if err := encoder.encodeAMF0(w, array[i]); err != nil {
			return err
		}
	}
	return nil
}

func (encoder *Encoder) EncodeDate(w io.Writer, date DateType) error {
	if err := encoder.EncodeMarker(w, DateMarker); err != nil {
		return err
	}
	d := math.Float64bits(float64(date))
	binary.BigEndian.PutUint64(encoder.u64[:], d)
	if _, err := w.Write(encoder.u64[:]); err != nil {
		return err
	}
	_, err := w.Write([]byte{0x00, 0x00}) // time zone
	return err
}

func (encoder *Encoder) EncodeLongString(w io.Writer, str string) error {
	length := len(str)
	if length > math.MaxUint32 {
		return errors.New("long string too long")
	}
	if err := encoder.EncodeMarker(w, LongStringMarker); err != nil {
		return err
	}
	return encoder.writeUTF8Long(w, str, uint32(length))
}

func (encoder *Encoder) EncodeXMLDocument(w io.Writer, document XMLDocumentType) error {
	if len(document) > math.MaxUint32 {
		return errors.New("xmlDocument too long")
	}
	if err := encoder.EncodeMarker(w, XMLDocumentMarker); err != nil {
		return err
	}
	return encoder.writeUTF8Long(w, string(document), uint32(len(document)))
}

func (encoder *Encoder) EncodeTypedObject(w io.Writer, object *TypedObjectType) error {
	if len(object.ClassName) > math.MaxUint16 {
		return errors.New("typedObject class name too long")
	}
	encoder.refObjects = append(encoder.refObjects, object)
	if err := encoder.EncodeMarker(w, TypedObjectMarker); err != nil {
		return err
	}
	if err := encoder.writeUTF8(w, object.ClassName, uint16(len(object.ClassName))); err != nil {
		return err
	}
	return encoder.writeObject(w, object.Object)
}

func (encoder *Encoder) writeObject(w io.Writer, m map[string]interface{}) error {
	if err := encoder.EncodeMarker(w, ObjectMarker); err != nil {
		return err
	}
	for k := range m {
		if len(k) > math.MaxUint16 {
			return errors.New("object key too long")
		}
		if err := encoder.writeUTF8(w, k, uint16(len(k))); err != nil {
			return err
		}
		if err := encoder.encodeAMF0(w, m[k]); err != nil {
			return err
		}
	}
	// k(u16) = 0, v = ObjectEndMarker
	encoder.u16[0] = 0
	encoder.u16[1] = 0
	if _, err := w.Write(encoder.u16[:]); err != nil {
		return err
	}
	return encoder.EncodeMarker(w, ObjectEndMarker)
}

func (encoder *Encoder) writeUTF8(w io.Writer, str string, length uint16) error {
	binary.BigEndian.PutUint16(encoder.u16[:], length)
	if _, err := w.Write(encoder.u16[:]); err != nil {
		return err
	}
	_, err := w.Write([]byte(str))
	return err
}

func (encoder *Encoder) writeUTF8Long(w io.Writer, str string, length uint32) error {
	binary.BigEndian.PutUint32(encoder.u32[:], length)
	if _, err := w.Write(encoder.u32[:]); err != nil {
		return err
	}
	_, err := w.Write([]byte(str))
	return err
}

func (encoder *Encoder) encodeAMF3(w io.Writer, v interface{}) error {
	switch t := v.(type) {
	case UndefinedType:
		return encoder.EncodeMarker(w, UndefinedMarkerAMF3)
	case NullType:
		return encoder.EncodeMarker(w, NullMarkerAMF3)
	case FalseTypeAMF3:
		return encoder.EncodeMarker(w, FalseMarkerAMF3)
	case TrueTypeAMF3:
		return encoder.EncodeMarker(w, TrueMarkerAMF3)
	case uint32:
		return encoder.EncodeIntegerAMF3(w, t)
	case float64:
		return encoder.EncodeNumberOrAMF3Double(w, t, DoubleMarkerAMF3)
	case string:
		return encoder.EncodeStringAMF3(w, t)
	case XMLDocumentType:
		return encoder.EncodeXMLDocAMF3(w, t)
	case DateType:
		return encoder.EncodeDateAMF3(w, t)
	case []interface{}: // array type, ignore associative
		return encoder.EncodeArrayAMF3(w, t)
	case *ObjectTypeAMF3:
		return encoder.EncodeObjectAMF3()
	case XMLTypeAMF3:
		return encoder.EncodeXMLAMF3(w, t)
	case []byte:
		return encoder.EncodeByteArrayAMF3(w, t)
	default:
		return fmt.Errorf("type %T not supported", t)
	}
}

func (encoder *Encoder) EncodeIntegerAMF3(w io.Writer, v uint32) error {
	if err := encoder.EncodeMarker(w, IntegerMarkerAMF3); err != nil {
		return err
	}
	return EncodeUint29(w, v)
}

func (encoder *Encoder) EncodeStringAMF3(w io.Writer, v string) error {
	if err := encoder.EncodeMarker(w, StringMarkerAMF3); err != nil {
		return err
	}
	return encoder.writeStringAMF3(w, v)
}

func (encoder *Encoder) EncodeXMLDocAMF3(w io.Writer, v XMLDocumentType) error {
	if err := encoder.EncodeMarker(w, XMLDocMarkerAMF3); err != nil {
		return err
	}
	if err := writeUTF8AMF3(w, string(v)); err != nil {
		return err
	}
	encoder.refObjects = append(encoder.refObjects, v)
	return nil
}

func (encoder *Encoder) EncodeDateAMF3(w io.Writer, v DateType) error {
	if err := encoder.EncodeMarker(w, DateMarkerAMF3); err != nil {
		return err
	}
	date := math.Float64bits(float64(v))
	binary.BigEndian.PutUint64(encoder.u64[:], date)
	if _, err := w.Write(encoder.u64[:]); err != nil {
		return err
	}
	encoder.refObjects = append(encoder.refObjects, v)
	return nil
}

func (encoder *Encoder) EncodeArrayAMF3(w io.Writer, v []interface{}) error {
	if err := encoder.EncodeMarker(w, ArrayMarkerAMF3); err != nil {
		return err
	}
	encoder.refObjects = append(encoder.refObjects, v)
	if err := EncodeUint29(w, uint32(len(v))<<1); err != nil {
		return err
	}
	if err := writeUTF8AMF3(w, ""); err != nil { // ignore associative
		return err
	}
	for i := range v {
		if err := encoder.encodeAMF3(w, v[i]); err != nil {
			return err
		}
	}
	return nil
}

func (encoder *Encoder) EncodeObjectAMF3() error {
	panic("amf3 encoder: not support object")
}

func (encoder *Encoder) EncodeXMLAMF3(w io.Writer, v XMLTypeAMF3) error {
	if err := encoder.EncodeMarker(w, XMLMarkerAMF3); err != nil {
		return err
	}
	if err := writeUTF8AMF3(w, string(v)); err != nil {
		return err
	}
	encoder.refObjects = append(encoder.refObjects, v)
	return nil
}

func (encoder *Encoder) EncodeByteArrayAMF3(w io.Writer, v []byte) error {
	if err := encoder.EncodeMarker(w, ByteArrayMarkerAMF3); err != nil {
		return err
	}
	length := len(v)
	if err := EncodeUint29(w, uint32(length<<1)); err != nil {
		return err
	}
	encoder.refObjects = append(encoder.refObjects, v)
	return nil
}

func (encoder *Encoder) writeStringAMF3(w io.Writer, str string) error {
	for i := range encoder.refStrings {
		if str == encoder.refStrings[i] {
			return EncodeUint29(w, uint32(i<<1|0x01))
		}
	}
	if err := writeUTF8AMF3(w, str); err != nil {
		return err
	}
	encoder.refStrings = append(encoder.refStrings, str)
	return nil
}

func writeUTF8AMF3(w io.Writer, str string) error {
	length := len(str)
	if err := EncodeUint29(w, uint32(length<<1)); err != nil {
		return err
	}
	_, err := w.Write([]byte(str))
	return err
}
