package amf

const (
	Version0 = 0
	Version3 = 3
)

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

type FalseTypeAMF3 struct{}
type TrueTypeAMF3 struct{}
type ArrayTypeAMF3 struct {
	Associative map[string]interface{}
	Dense       []interface{}
}
type TraitAMF3 struct {
	ClassName  string
	IsDynamic  bool
	Attributes []string
}
type ObjectTypeAMF3 struct {
	Trait   *TraitAMF3
	Static  []interface{}
	Dynamic map[string]interface{}
}
type XMLTypeAMF3 string

// amf0
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

// amf3
const (
	UndefinedMarkerAMF3 byte = iota
	NullMarkerAMF3
	FalseMarkerAMF3
	TrueMarkerAMF3
	IntegerMarkerAMF3
	DoubleMarkerAMF3
	StringMarkerAMF3
	XMLDocMarkerAMF3
	DateMarkerAMF3
	ArrayMarkerAMF3
	ObjectMarkerAMF3
	XMLMarkerAMF3
	ByteArrayMarkerAMF3
	VectorIntMarkerAMF3
	VectorUintMarkerAMF3
	VectorDoubleMarkerAMF3
	VectorObjectMarkerAMF3
	DictionaryMarkerAMF3
)
