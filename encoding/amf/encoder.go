package amf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

func NewEncoder() *Encoder {
	return &Encoder{
		refObjects: make([]interface{}, 0),
	}
}

type Encoder struct {
	u8         [1]byte
	u16        [2]byte
	u32        [4]byte
	u64        [8]byte
	refObjects []interface{}
}

func (encoder *Encoder) Encode(w io.Writer, v interface{}) error {
	return encoder.encode(w, v)
}

func (encoder *Encoder) EncodeBatch(w io.Writer, vs ...interface{}) error {
	for i := range vs {
		if err := encoder.encode(w, vs[i]); err != nil {
			return err
		}
	}
	return nil
}

func (encoder *Encoder) encode(w io.Writer, v interface{}) error {
	switch t := v.(type) {
	case float64: // number type
		return encoder.EncodeNumber(w, t)
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

func (encoder *Encoder) EncodeNumber(w io.Writer, v float64) error {
	if err := encoder.EncodeMarker(w, NumberMarker); err != nil {
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
	ok, err := encoder.writeReference(w, m)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	encoder.refObjects = append(encoder.refObjects, m)
	return encoder.writeObject(w, m)
}

func (encoder *Encoder) EncodeECMAArray(w io.Writer, array ECMAArray) error {
	ok, err := encoder.writeReference(w, array)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
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
	ok, err := encoder.writeReference(w, array)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
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
		if err := encoder.encode(w, array[i]); err != nil {
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
	ok, err := encoder.writeReference(w, object)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
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
		if err := encoder.encode(w, m[k]); err != nil {
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

func (encoder *Encoder) writeReference(w io.Writer, v interface{}) (bool, error) {
	for i := range encoder.refObjects {
		if encoder.refObjects[i] == v {
			if err := encoder.EncodeMarker(w, ReferenceMarker); err != nil {
				return false, err
			}
			binary.BigEndian.PutUint16(encoder.u16[:], uint16(i))
			if _, err := w.Write(encoder.u16[:]); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
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
