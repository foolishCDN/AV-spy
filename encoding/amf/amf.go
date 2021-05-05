package amf

type NullType struct{}
type UndefinedType struct{}
type UnsupportedType struct{}
type DateType float64
type XMLDocumentType string
type TypedObjectType struct {
	ClassName string
	Object    map[string]interface{}
}
type ECMAArray map[string]interface{}

const (
	NumberMarker byte = iota
	BooleanMarker
	StringMarker
	ObjectMarker
	MovieClipMarker
	NullMarker
	UndefinedMarker
	ReferenceMarker
	ECMAArrayMarker
	ObjectEndMarker
	StrictArrayMarker
	DateMarker
	LongStringMarker
	UnSupportedMarker
	RecordSetMarker
	XMLDocumentMarker
	TypedObjectMarker
	AVMPlusObjectMarker // signal a switch to AMF 3 serialization
)
