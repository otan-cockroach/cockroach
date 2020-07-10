// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: geo/geopb/geopb.proto

package geopb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import encoding_binary "encoding/binary"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// ShapeType is the type of a spatial shape. Each of these corresponds to a
// different representation and serialization format. For example, a Point is a
// pair of doubles (or more than that for geometries with Z or N), a LineString
// is an ordered series of Points, etc.
type ShapeType int32

const (
	ShapeType_Unset           ShapeType = 0
	ShapeType_Point           ShapeType = 1
	ShapeType_LineString      ShapeType = 2
	ShapeType_Polygon         ShapeType = 3
	ShapeType_MultiPoint      ShapeType = 4
	ShapeType_MultiLineString ShapeType = 5
	ShapeType_MultiPolygon    ShapeType = 6
	// Geometry can contain any type.
	ShapeType_Geometry ShapeType = 7
	// GeometryCollection can contain a list of any above type except for Geometry.
	ShapeType_GeometryCollection ShapeType = 8
)

var ShapeType_name = map[int32]string{
	0: "Unset",
	1: "Point",
	2: "LineString",
	3: "Polygon",
	4: "MultiPoint",
	5: "MultiLineString",
	6: "MultiPolygon",
	7: "Geometry",
	8: "GeometryCollection",
}
var ShapeType_value = map[string]int32{
	"Unset":              0,
	"Point":              1,
	"LineString":         2,
	"Polygon":            3,
	"MultiPoint":         4,
	"MultiLineString":    5,
	"MultiPolygon":       6,
	"Geometry":           7,
	"GeometryCollection": 8,
}

func (x ShapeType) String() string {
	return proto.EnumName(ShapeType_name, int32(x))
}
func (ShapeType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_geopb_dec0be5909ed5664, []int{0}
}

// SpatialObjectType represents the type of the SpatialObject.
type SpatialObjectType int32

const (
	SpatialObjectType_Unknown       SpatialObjectType = 0
	SpatialObjectType_GeographyType SpatialObjectType = 1
	SpatialObjectType_GeometryType  SpatialObjectType = 2
)

var SpatialObjectType_name = map[int32]string{
	0: "Unknown",
	1: "GeographyType",
	2: "GeometryType",
}
var SpatialObjectType_value = map[string]int32{
	"Unknown":       0,
	"GeographyType": 1,
	"GeometryType":  2,
}

func (x SpatialObjectType) String() string {
	return proto.EnumName(SpatialObjectType_name, int32(x))
}
func (SpatialObjectType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_geopb_dec0be5909ed5664, []int{1}
}

// Shape represents a single shape.
type Shape struct {
	// Coords is a flat representation of individual components of the SpatialObject.
	// They may be split up by the Ends variable.
	Coords []float64 `protobuf:"fixed64,1,rep,packed,name=coords,proto3" json:"coords,omitempty"`
	// Ends denotes when to split up coords array.
	// For example, for Polygons, there may be multiple rings. Ends is the index to slice
	// at to get each ring.
	Ends []int `protobuf:"varint,2,rep,packed,name=ends,proto3,casttype=int" json:"ends,omitempty"`
	// EndsEnds denotes when to split up the ends array.
	// For example, for MultiPolygons, there are multiple ends for multiple rings.
	// EndsEnds is the index to slice at of the last element of each polygon.
	EndsEnds []int `protobuf:"varint,3,rep,packed,name=ends_ends,json=endsEnds,proto3,casttype=int" json:"ends_ends,omitempty"`
}

func (m *Shape) Reset()         { *m = Shape{} }
func (m *Shape) String() string { return proto.CompactTextString(m) }
func (*Shape) ProtoMessage()    {}
func (*Shape) Descriptor() ([]byte, []int) {
	return fileDescriptor_geopb_dec0be5909ed5664, []int{0}
}
func (m *Shape) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Shape) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *Shape) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Shape.Merge(dst, src)
}
func (m *Shape) XXX_Size() int {
	return m.Size()
}
func (m *Shape) XXX_DiscardUnknown() {
	xxx_messageInfo_Shape.DiscardUnknown(m)
}

var xxx_messageInfo_Shape proto.InternalMessageInfo

// GeometryCollectionShape represents a GeometryCollection shape.
type GeometryCollectionShape struct {
	// Shapes represents all the shapes for a GeometryCollection.
	Shapes []GeometryCollectionShape_GeometryCollectionSubShape `protobuf:"bytes,1,rep,name=shapes,proto3" json:"shapes"`
}

func (m *GeometryCollectionShape) Reset()         { *m = GeometryCollectionShape{} }
func (m *GeometryCollectionShape) String() string { return proto.CompactTextString(m) }
func (*GeometryCollectionShape) ProtoMessage()    {}
func (*GeometryCollectionShape) Descriptor() ([]byte, []int) {
	return fileDescriptor_geopb_dec0be5909ed5664, []int{1}
}
func (m *GeometryCollectionShape) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GeometryCollectionShape) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *GeometryCollectionShape) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GeometryCollectionShape.Merge(dst, src)
}
func (m *GeometryCollectionShape) XXX_Size() int {
	return m.Size()
}
func (m *GeometryCollectionShape) XXX_DiscardUnknown() {
	xxx_messageInfo_GeometryCollectionShape.DiscardUnknown(m)
}

var xxx_messageInfo_GeometryCollectionShape proto.InternalMessageInfo

type GeometryCollectionShape_GeometryCollectionSubShape struct {
	Shape     Shape     `protobuf:"bytes,1,opt,name=shape,proto3" json:"shape"`
	ShapeType ShapeType `protobuf:"varint,2,opt,name=shape_type,json=shapeType,proto3,enum=cockroach.geopb.ShapeType" json:"shape_type,omitempty"`
}

func (m *GeometryCollectionShape_GeometryCollectionSubShape) Reset() {
	*m = GeometryCollectionShape_GeometryCollectionSubShape{}
}
func (m *GeometryCollectionShape_GeometryCollectionSubShape) String() string {
	return proto.CompactTextString(m)
}
func (*GeometryCollectionShape_GeometryCollectionSubShape) ProtoMessage() {}
func (*GeometryCollectionShape_GeometryCollectionSubShape) Descriptor() ([]byte, []int) {
	return fileDescriptor_geopb_dec0be5909ed5664, []int{1, 0}
}
func (m *GeometryCollectionShape_GeometryCollectionSubShape) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GeometryCollectionShape_GeometryCollectionSubShape) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *GeometryCollectionShape_GeometryCollectionSubShape) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GeometryCollectionShape_GeometryCollectionSubShape.Merge(dst, src)
}
func (m *GeometryCollectionShape_GeometryCollectionSubShape) XXX_Size() int {
	return m.Size()
}
func (m *GeometryCollectionShape_GeometryCollectionSubShape) XXX_DiscardUnknown() {
	xxx_messageInfo_GeometryCollectionShape_GeometryCollectionSubShape.DiscardUnknown(m)
}

var xxx_messageInfo_GeometryCollectionShape_GeometryCollectionSubShape proto.InternalMessageInfo

// SpatialObject represents a serialization of a Geospatial type.
type SpatialObject struct {
	// Type is the type of the SpatialObject.
	Type SpatialObjectType `protobuf:"varint,1,opt,name=type,proto3,enum=cockroach.geopb.SpatialObjectType" json:"type,omitempty"`
	// SerializedShape is the serialized Shape or GeometryCollection representation of the Shape.
	// This is not included to avoid the potentially expensive deserialization of the Shape.
	SerializedShape []byte `protobuf:"bytes,2,opt,name=serialized_shape,json=serializedShape,proto3" json:"serialized_shape,omitempty"`
	// SRID is the denormalized SRID derived from the EWKB.
	SRID SRID `protobuf:"varint,3,opt,name=srid,proto3,casttype=SRID" json:"srid,omitempty"`
	// ShapeType is denormalized ShapeType derived from the EWKB.
	ShapeType ShapeType `protobuf:"varint,4,opt,name=shape_type,json=shapeType,proto3,enum=cockroach.geopb.ShapeType" json:"shape_type,omitempty"`
	// BoundingBox is the bounding box of the SpatialObject.
	BoundingBox *BoundingBox `protobuf:"bytes,5,opt,name=bounding_box,json=boundingBox,proto3" json:"bounding_box,omitempty"`
}

func (m *SpatialObject) Reset()         { *m = SpatialObject{} }
func (m *SpatialObject) String() string { return proto.CompactTextString(m) }
func (*SpatialObject) ProtoMessage()    {}
func (*SpatialObject) Descriptor() ([]byte, []int) {
	return fileDescriptor_geopb_dec0be5909ed5664, []int{2}
}
func (m *SpatialObject) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SpatialObject) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *SpatialObject) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpatialObject.Merge(dst, src)
}
func (m *SpatialObject) XXX_Size() int {
	return m.Size()
}
func (m *SpatialObject) XXX_DiscardUnknown() {
	xxx_messageInfo_SpatialObject.DiscardUnknown(m)
}

var xxx_messageInfo_SpatialObject proto.InternalMessageInfo

// BoundingBox represents the bounding box of a Geospatial type.
// Note the lo coordinates can be higher in value than the hi coordinates
// for spherical geometries.
// NOTE: Do not use these to compare bounding boxes. Use the library functions
// provided in the geo package to perform these calculations.
type BoundingBox struct {
	LoX float64 `protobuf:"fixed64,1,opt,name=lo_x,json=loX,proto3" json:"lo_x,omitempty"`
	HiX float64 `protobuf:"fixed64,2,opt,name=hi_x,json=hiX,proto3" json:"hi_x,omitempty"`
	LoY float64 `protobuf:"fixed64,3,opt,name=lo_y,json=loY,proto3" json:"lo_y,omitempty"`
	HiY float64 `protobuf:"fixed64,4,opt,name=hi_y,json=hiY,proto3" json:"hi_y,omitempty"`
}

func (m *BoundingBox) Reset()         { *m = BoundingBox{} }
func (m *BoundingBox) String() string { return proto.CompactTextString(m) }
func (*BoundingBox) ProtoMessage()    {}
func (*BoundingBox) Descriptor() ([]byte, []int) {
	return fileDescriptor_geopb_dec0be5909ed5664, []int{3}
}
func (m *BoundingBox) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BoundingBox) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *BoundingBox) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BoundingBox.Merge(dst, src)
}
func (m *BoundingBox) XXX_Size() int {
	return m.Size()
}
func (m *BoundingBox) XXX_DiscardUnknown() {
	xxx_messageInfo_BoundingBox.DiscardUnknown(m)
}

var xxx_messageInfo_BoundingBox proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Shape)(nil), "cockroach.geopb.Shape")
	proto.RegisterType((*GeometryCollectionShape)(nil), "cockroach.geopb.GeometryCollectionShape")
	proto.RegisterType((*GeometryCollectionShape_GeometryCollectionSubShape)(nil), "cockroach.geopb.GeometryCollectionShape.GeometryCollectionSubShape")
	proto.RegisterType((*SpatialObject)(nil), "cockroach.geopb.SpatialObject")
	proto.RegisterType((*BoundingBox)(nil), "cockroach.geopb.BoundingBox")
	proto.RegisterEnum("cockroach.geopb.ShapeType", ShapeType_name, ShapeType_value)
	proto.RegisterEnum("cockroach.geopb.SpatialObjectType", SpatialObjectType_name, SpatialObjectType_value)
}
func (m *Shape) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Shape) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Coords) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(len(m.Coords)*8))
		for _, num := range m.Coords {
			f1 := math.Float64bits(float64(num))
			encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(f1))
			i += 8
		}
	}
	if len(m.Ends) > 0 {
		dAtA3 := make([]byte, len(m.Ends)*10)
		var j2 int
		for _, num1 := range m.Ends {
			num := uint64(num1)
			for num >= 1<<7 {
				dAtA3[j2] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j2++
			}
			dAtA3[j2] = uint8(num)
			j2++
		}
		dAtA[i] = 0x12
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(j2))
		i += copy(dAtA[i:], dAtA3[:j2])
	}
	if len(m.EndsEnds) > 0 {
		dAtA5 := make([]byte, len(m.EndsEnds)*10)
		var j4 int
		for _, num1 := range m.EndsEnds {
			num := uint64(num1)
			for num >= 1<<7 {
				dAtA5[j4] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j4++
			}
			dAtA5[j4] = uint8(num)
			j4++
		}
		dAtA[i] = 0x1a
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(j4))
		i += copy(dAtA[i:], dAtA5[:j4])
	}
	return i, nil
}

func (m *GeometryCollectionShape) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GeometryCollectionShape) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Shapes) > 0 {
		for _, msg := range m.Shapes {
			dAtA[i] = 0xa
			i++
			i = encodeVarintGeopb(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *GeometryCollectionShape_GeometryCollectionSubShape) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GeometryCollectionShape_GeometryCollectionSubShape) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintGeopb(dAtA, i, uint64(m.Shape.Size()))
	n6, err := m.Shape.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n6
	if m.ShapeType != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(m.ShapeType))
	}
	return i, nil
}

func (m *SpatialObject) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SpatialObject) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(m.Type))
	}
	if len(m.SerializedShape) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(len(m.SerializedShape)))
		i += copy(dAtA[i:], m.SerializedShape)
	}
	if m.SRID != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(m.SRID))
	}
	if m.ShapeType != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(m.ShapeType))
	}
	if m.BoundingBox != nil {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintGeopb(dAtA, i, uint64(m.BoundingBox.Size()))
		n7, err := m.BoundingBox.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n7
	}
	return i, nil
}

func (m *BoundingBox) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BoundingBox) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.LoX != 0 {
		dAtA[i] = 0x9
		i++
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.LoX))))
		i += 8
	}
	if m.HiX != 0 {
		dAtA[i] = 0x11
		i++
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.HiX))))
		i += 8
	}
	if m.LoY != 0 {
		dAtA[i] = 0x19
		i++
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.LoY))))
		i += 8
	}
	if m.HiY != 0 {
		dAtA[i] = 0x21
		i++
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.HiY))))
		i += 8
	}
	return i, nil
}

func encodeVarintGeopb(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Shape) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Coords) > 0 {
		n += 1 + sovGeopb(uint64(len(m.Coords)*8)) + len(m.Coords)*8
	}
	if len(m.Ends) > 0 {
		l = 0
		for _, e := range m.Ends {
			l += sovGeopb(uint64(e))
		}
		n += 1 + sovGeopb(uint64(l)) + l
	}
	if len(m.EndsEnds) > 0 {
		l = 0
		for _, e := range m.EndsEnds {
			l += sovGeopb(uint64(e))
		}
		n += 1 + sovGeopb(uint64(l)) + l
	}
	return n
}

func (m *GeometryCollectionShape) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Shapes) > 0 {
		for _, e := range m.Shapes {
			l = e.Size()
			n += 1 + l + sovGeopb(uint64(l))
		}
	}
	return n
}

func (m *GeometryCollectionShape_GeometryCollectionSubShape) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Shape.Size()
	n += 1 + l + sovGeopb(uint64(l))
	if m.ShapeType != 0 {
		n += 1 + sovGeopb(uint64(m.ShapeType))
	}
	return n
}

func (m *SpatialObject) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovGeopb(uint64(m.Type))
	}
	l = len(m.SerializedShape)
	if l > 0 {
		n += 1 + l + sovGeopb(uint64(l))
	}
	if m.SRID != 0 {
		n += 1 + sovGeopb(uint64(m.SRID))
	}
	if m.ShapeType != 0 {
		n += 1 + sovGeopb(uint64(m.ShapeType))
	}
	if m.BoundingBox != nil {
		l = m.BoundingBox.Size()
		n += 1 + l + sovGeopb(uint64(l))
	}
	return n
}

func (m *BoundingBox) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.LoX != 0 {
		n += 9
	}
	if m.HiX != 0 {
		n += 9
	}
	if m.LoY != 0 {
		n += 9
	}
	if m.HiY != 0 {
		n += 9
	}
	return n
}

func sovGeopb(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozGeopb(x uint64) (n int) {
	return sovGeopb(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Shape) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGeopb
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Shape: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Shape: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 1 {
				var v uint64
				if (iNdEx + 8) > l {
					return io.ErrUnexpectedEOF
				}
				v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
				iNdEx += 8
				v2 := float64(math.Float64frombits(v))
				m.Coords = append(m.Coords, v2)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGeopb
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthGeopb
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				elementCount = packedLen / 8
				if elementCount != 0 && len(m.Coords) == 0 {
					m.Coords = make([]float64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint64
					if (iNdEx + 8) > l {
						return io.ErrUnexpectedEOF
					}
					v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
					iNdEx += 8
					v2 := float64(math.Float64frombits(v))
					m.Coords = append(m.Coords, v2)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Coords", wireType)
			}
		case 2:
			if wireType == 0 {
				var v int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGeopb
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Ends = append(m.Ends, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGeopb
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthGeopb
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.Ends) == 0 {
					m.Ends = make([]int, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGeopb
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= (int(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Ends = append(m.Ends, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Ends", wireType)
			}
		case 3:
			if wireType == 0 {
				var v int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGeopb
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.EndsEnds = append(m.EndsEnds, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGeopb
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthGeopb
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.EndsEnds) == 0 {
					m.EndsEnds = make([]int, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGeopb
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= (int(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.EndsEnds = append(m.EndsEnds, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field EndsEnds", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGeopb(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGeopb
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GeometryCollectionShape) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGeopb
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GeometryCollectionShape: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GeometryCollectionShape: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Shapes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGeopb
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Shapes = append(m.Shapes, GeometryCollectionShape_GeometryCollectionSubShape{})
			if err := m.Shapes[len(m.Shapes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGeopb(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGeopb
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GeometryCollectionShape_GeometryCollectionSubShape) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGeopb
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GeometryCollectionSubShape: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GeometryCollectionSubShape: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Shape", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGeopb
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Shape.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ShapeType", wireType)
			}
			m.ShapeType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ShapeType |= (ShapeType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGeopb(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGeopb
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SpatialObject) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGeopb
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SpatialObject: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SpatialObject: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (SpatialObjectType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SerializedShape", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGeopb
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SerializedShape = append(m.SerializedShape[:0], dAtA[iNdEx:postIndex]...)
			if m.SerializedShape == nil {
				m.SerializedShape = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SRID", wireType)
			}
			m.SRID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SRID |= (SRID(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ShapeType", wireType)
			}
			m.ShapeType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ShapeType |= (ShapeType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BoundingBox", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGeopb
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.BoundingBox == nil {
				m.BoundingBox = &BoundingBox{}
			}
			if err := m.BoundingBox.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGeopb(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGeopb
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *BoundingBox) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGeopb
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BoundingBox: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BoundingBox: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field LoX", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.LoX = float64(math.Float64frombits(v))
		case 2:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field HiX", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.HiX = float64(math.Float64frombits(v))
		case 3:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field LoY", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.LoY = float64(math.Float64frombits(v))
		case 4:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field HiY", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.HiY = float64(math.Float64frombits(v))
		default:
			iNdEx = preIndex
			skippy, err := skipGeopb(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGeopb
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGeopb(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGeopb
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGeopb
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthGeopb
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowGeopb
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipGeopb(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthGeopb = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGeopb   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("geo/geopb/geopb.proto", fileDescriptor_geopb_dec0be5909ed5664) }

var fileDescriptor_geopb_dec0be5909ed5664 = []byte{
	// 591 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xc7, 0xbd, 0xfe, 0x48, 0x9b, 0x49, 0xda, 0xba, 0x0b, 0x14, 0x2b, 0x20, 0x37, 0x8a, 0x2a,
	0x11, 0x7a, 0x48, 0xa5, 0x20, 0x21, 0x71, 0x42, 0x72, 0xa9, 0x2a, 0x24, 0x10, 0x95, 0x43, 0xa5,
	0x96, 0x8b, 0xe5, 0x8f, 0x95, 0xb3, 0xd4, 0xec, 0x5a, 0xb6, 0x2b, 0x62, 0x1e, 0x01, 0x2e, 0x1c,
	0xb8, 0xf1, 0x42, 0x39, 0xf6, 0xd8, 0x53, 0x04, 0xc9, 0x5b, 0xf4, 0x84, 0xbc, 0x76, 0x68, 0x68,
	0xda, 0x03, 0x17, 0xef, 0xec, 0xcc, 0x6f, 0xfe, 0x33, 0xb3, 0xbb, 0x86, 0x07, 0x21, 0xe1, 0x7b,
	0x21, 0xe1, 0xb1, 0x57, 0x7e, 0x7b, 0x71, 0xc2, 0x33, 0x8e, 0x37, 0x7c, 0xee, 0x9f, 0x25, 0xdc,
	0xf5, 0x87, 0x3d, 0xe1, 0x6e, 0xdd, 0x0f, 0x79, 0xc8, 0x45, 0x6c, 0xaf, 0xb0, 0x4a, 0xac, 0xe3,
	0x81, 0x36, 0x18, 0xba, 0x31, 0xc1, 0x5b, 0x50, 0xf3, 0x39, 0x4f, 0x82, 0xd4, 0x40, 0x6d, 0xa5,
	0x8b, 0xec, 0x6a, 0x87, 0x1f, 0x81, 0x4a, 0x58, 0x90, 0x1a, 0x72, 0x5b, 0xe9, 0x2a, 0xd6, 0xca,
	0xd5, 0x64, 0x5b, 0xa1, 0x2c, 0xb3, 0x85, 0x13, 0xef, 0x40, 0xbd, 0x58, 0x1d, 0x41, 0x28, 0xff,
	0x12, 0xab, 0x85, 0xf3, 0x80, 0x05, 0x69, 0xe7, 0xab, 0x0c, 0x0f, 0x0f, 0x09, 0xff, 0x44, 0xb2,
	0x24, 0xdf, 0xe7, 0x51, 0x44, 0xfc, 0x8c, 0x72, 0x56, 0x96, 0x75, 0xa1, 0x96, 0x16, 0x46, 0x59,
	0xb6, 0xd1, 0xdf, 0xef, 0xdd, 0xe8, 0xbb, 0x77, 0x47, 0xe6, 0x6d, 0xfe, 0x73, 0x4f, 0x84, 0x2c,
	0x75, 0x3c, 0xd9, 0x96, 0xec, 0x4a, 0xb8, 0xf5, 0x0d, 0x41, 0xeb, 0x6e, 0x18, 0xf7, 0x41, 0x13,
	0xa0, 0x81, 0xda, 0xa8, 0xdb, 0xe8, 0x6f, 0x2d, 0x35, 0xb0, 0xa8, 0x59, 0xa2, 0xf8, 0x05, 0x80,
	0x30, 0x9c, 0x2c, 0x8f, 0x89, 0x21, 0xb7, 0x51, 0x77, 0xbd, 0xdf, 0xba, 0x3d, 0xf1, 0x7d, 0x1e,
	0x13, 0xbb, 0x9e, 0xce, 0xcd, 0xce, 0x0f, 0x19, 0xd6, 0x06, 0xb1, 0x9b, 0x51, 0x37, 0x7a, 0xe7,
	0x7d, 0x24, 0x7e, 0x86, 0x9f, 0x83, 0x2a, 0x64, 0x90, 0x90, 0xe9, 0x2c, 0xcb, 0x2c, 0xd2, 0x42,
	0x4e, 0xf0, 0xf8, 0x29, 0xe8, 0x29, 0x49, 0xa8, 0x1b, 0xd1, 0x2f, 0x24, 0x70, 0xca, 0x19, 0x8a,
	0x56, 0x9a, 0xf6, 0xc6, 0xb5, 0xbf, 0x9c, 0x71, 0x07, 0xd4, 0x34, 0xa1, 0x81, 0xa1, 0xb4, 0x51,
	0x57, 0xb3, 0xf4, 0xe9, 0x64, 0x5b, 0x1d, 0xd8, 0xaf, 0x5f, 0x5d, 0x55, 0xab, 0x2d, 0xa2, 0x37,
	0xa6, 0x52, 0xff, 0x63, 0x2a, 0xfc, 0x12, 0x9a, 0x1e, 0x3f, 0x67, 0x01, 0x65, 0xa1, 0xe3, 0xf1,
	0x91, 0xa1, 0x89, 0xb3, 0x7c, 0xbc, 0x94, 0x6c, 0x55, 0x90, 0xc5, 0x47, 0x76, 0xc3, 0xbb, 0xde,
	0x74, 0x4e, 0xa1, 0xb1, 0x10, 0xc3, 0x9b, 0xa0, 0x46, 0xdc, 0x19, 0x89, 0x33, 0x41, 0xb6, 0x12,
	0xf1, 0x93, 0xc2, 0x35, 0xa4, 0xce, 0x48, 0x8c, 0x88, 0x6c, 0x65, 0x48, 0x4f, 0x2a, 0x2a, 0x17,
	0x63, 0x09, 0xea, 0xb4, 0xa2, 0x72, 0xd1, 0xbd, 0xa0, 0x4e, 0x77, 0x7f, 0x22, 0xa8, 0xff, 0x6d,
	0x1a, 0xd7, 0x41, 0x3b, 0x66, 0x29, 0xc9, 0x74, 0xa9, 0x30, 0x8f, 0x38, 0x65, 0x99, 0x8e, 0xf0,
	0x3a, 0xc0, 0x1b, 0xca, 0xc8, 0x20, 0x4b, 0x28, 0x0b, 0x75, 0x19, 0x37, 0x60, 0xe5, 0x88, 0x47,
	0x79, 0xc8, 0x99, 0xae, 0x14, 0xc1, 0xb7, 0xe7, 0x51, 0x46, 0x4b, 0x58, 0xc5, 0xf7, 0x60, 0x43,
	0xec, 0x17, 0x32, 0x34, 0xac, 0x43, 0xb3, 0x82, 0xca, 0xb4, 0x1a, 0x6e, 0xc2, 0xea, 0xfc, 0xd9,
	0xe9, 0x2b, 0x78, 0x0b, 0xf0, 0xf2, 0x23, 0xd4, 0x57, 0x77, 0x0f, 0x60, 0x73, 0xe9, 0x82, 0x8b,
	0xf2, 0xc7, 0xec, 0x8c, 0xf1, 0xcf, 0x4c, 0x97, 0xf0, 0x26, 0xac, 0x1d, 0x12, 0x1e, 0x26, 0x6e,
	0x3c, 0xcc, 0x8b, 0xa8, 0x8e, 0x8a, 0x62, 0x73, 0x31, 0xe1, 0x91, 0xad, 0x27, 0xe3, 0xdf, 0xa6,
	0x34, 0x9e, 0x9a, 0xe8, 0x62, 0x6a, 0xa2, 0xcb, 0xa9, 0x89, 0x7e, 0x4d, 0x4d, 0xf4, 0x7d, 0x66,
	0x4a, 0x17, 0x33, 0x53, 0xba, 0x9c, 0x99, 0xd2, 0x07, 0x4d, 0xdc, 0x80, 0x57, 0x13, 0xff, 0xfd,
	0xb3, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x2f, 0x47, 0xb8, 0x11, 0x37, 0x04, 0x00, 0x00,
}
