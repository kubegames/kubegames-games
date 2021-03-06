// Code generated by protoc-gen-go. DO NOT EDIT.
// source: game_c2s.proto

package msg

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type C2SMsgType int32

const (
	C2SMsgType_ROOM_INFO         C2SMsgType = 0
	C2SMsgType_START_MATCH       C2SMsgType = 1
	C2SMsgType_SET_CARDS         C2SMsgType = 2
	C2SMsgType_USER_SELECT_CARDS C2SMsgType = 3
)

var C2SMsgType_name = map[int32]string{
	0: "ROOM_INFO",
	1: "START_MATCH",
	2: "SET_CARDS",
	3: "USER_SELECT_CARDS",
}

var C2SMsgType_value = map[string]int32{
	"ROOM_INFO":         0,
	"START_MATCH":       1,
	"SET_CARDS":         2,
	"USER_SELECT_CARDS": 3,
}

func (x C2SMsgType) String() string {
	return proto.EnumName(C2SMsgType_name, int32(x))
}

func (C2SMsgType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_335bee2329d289c2, []int{0}
}

//获取房间信息
type C2SRoomInfo struct {
	Uid                  int64    `protobuf:"varint,1,opt,name=uid,proto3" json:"uid"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *C2SRoomInfo) Reset()         { *m = C2SRoomInfo{} }
func (m *C2SRoomInfo) String() string { return proto.CompactTextString(m) }
func (*C2SRoomInfo) ProtoMessage()    {}
func (*C2SRoomInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_335bee2329d289c2, []int{0}
}

func (m *C2SRoomInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_C2SRoomInfo.Unmarshal(m, b)
}
func (m *C2SRoomInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_C2SRoomInfo.Marshal(b, m, deterministic)
}
func (m *C2SRoomInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_C2SRoomInfo.Merge(m, src)
}
func (m *C2SRoomInfo) XXX_Size() int {
	return xxx_messageInfo_C2SRoomInfo.Size(m)
}
func (m *C2SRoomInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_C2SRoomInfo.DiscardUnknown(m)
}

var xxx_messageInfo_C2SRoomInfo proto.InternalMessageInfo

func (m *C2SRoomInfo) GetUid() int64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

//开始匹配
type C2SStartMatch struct {
	Uid                  int64    `protobuf:"varint,1,opt,name=uid,proto3" json:"uid"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *C2SStartMatch) Reset()         { *m = C2SStartMatch{} }
func (m *C2SStartMatch) String() string { return proto.CompactTextString(m) }
func (*C2SStartMatch) ProtoMessage()    {}
func (*C2SStartMatch) Descriptor() ([]byte, []int) {
	return fileDescriptor_335bee2329d289c2, []int{1}
}

func (m *C2SStartMatch) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_C2SStartMatch.Unmarshal(m, b)
}
func (m *C2SStartMatch) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_C2SStartMatch.Marshal(b, m, deterministic)
}
func (m *C2SStartMatch) XXX_Merge(src proto.Message) {
	xxx_messageInfo_C2SStartMatch.Merge(m, src)
}
func (m *C2SStartMatch) XXX_Size() int {
	return xxx_messageInfo_C2SStartMatch.Size(m)
}
func (m *C2SStartMatch) XXX_DiscardUnknown() {
	xxx_messageInfo_C2SStartMatch.DiscardUnknown(m)
}

var xxx_messageInfo_C2SStartMatch proto.InternalMessageInfo

func (m *C2SStartMatch) GetUid() int64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

//手动摆牌
type C2SSetCards struct {
	HeadCards            []byte   `protobuf:"bytes,1,opt,name=headCards,proto3" json:"headCards"`
	MidCards             []byte   `protobuf:"bytes,2,opt,name=midCards,proto3" json:"midCards"`
	TailCards            []byte   `protobuf:"bytes,3,opt,name=tailCards,proto3" json:"tailCards"`
	IsAuto               bool     `protobuf:"varint,4,opt,name=isAuto,proto3" json:"isAuto"`
	IsSpecial            bool     `protobuf:"varint,5,opt,name=isSpecial,proto3" json:"isSpecial"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *C2SSetCards) Reset()         { *m = C2SSetCards{} }
func (m *C2SSetCards) String() string { return proto.CompactTextString(m) }
func (*C2SSetCards) ProtoMessage()    {}
func (*C2SSetCards) Descriptor() ([]byte, []int) {
	return fileDescriptor_335bee2329d289c2, []int{2}
}

func (m *C2SSetCards) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_C2SSetCards.Unmarshal(m, b)
}
func (m *C2SSetCards) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_C2SSetCards.Marshal(b, m, deterministic)
}
func (m *C2SSetCards) XXX_Merge(src proto.Message) {
	xxx_messageInfo_C2SSetCards.Merge(m, src)
}
func (m *C2SSetCards) XXX_Size() int {
	return xxx_messageInfo_C2SSetCards.Size(m)
}
func (m *C2SSetCards) XXX_DiscardUnknown() {
	xxx_messageInfo_C2SSetCards.DiscardUnknown(m)
}

var xxx_messageInfo_C2SSetCards proto.InternalMessageInfo

func (m *C2SSetCards) GetHeadCards() []byte {
	if m != nil {
		return m.HeadCards
	}
	return nil
}

func (m *C2SSetCards) GetMidCards() []byte {
	if m != nil {
		return m.MidCards
	}
	return nil
}

func (m *C2SSetCards) GetTailCards() []byte {
	if m != nil {
		return m.TailCards
	}
	return nil
}

func (m *C2SSetCards) GetIsAuto() bool {
	if m != nil {
		return m.IsAuto
	}
	return false
}

func (m *C2SSetCards) GetIsSpecial() bool {
	if m != nil {
		return m.IsSpecial
	}
	return false
}

//
type C2SUserSelectCards struct {
	Cards                []byte   `protobuf:"bytes,1,opt,name=cards,proto3" json:"cards"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *C2SUserSelectCards) Reset()         { *m = C2SUserSelectCards{} }
func (m *C2SUserSelectCards) String() string { return proto.CompactTextString(m) }
func (*C2SUserSelectCards) ProtoMessage()    {}
func (*C2SUserSelectCards) Descriptor() ([]byte, []int) {
	return fileDescriptor_335bee2329d289c2, []int{3}
}

func (m *C2SUserSelectCards) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_C2SUserSelectCards.Unmarshal(m, b)
}
func (m *C2SUserSelectCards) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_C2SUserSelectCards.Marshal(b, m, deterministic)
}
func (m *C2SUserSelectCards) XXX_Merge(src proto.Message) {
	xxx_messageInfo_C2SUserSelectCards.Merge(m, src)
}
func (m *C2SUserSelectCards) XXX_Size() int {
	return xxx_messageInfo_C2SUserSelectCards.Size(m)
}
func (m *C2SUserSelectCards) XXX_DiscardUnknown() {
	xxx_messageInfo_C2SUserSelectCards.DiscardUnknown(m)
}

var xxx_messageInfo_C2SUserSelectCards proto.InternalMessageInfo

func (m *C2SUserSelectCards) GetCards() []byte {
	if m != nil {
		return m.Cards
	}
	return nil
}

func init() {
	proto.RegisterEnum("msg.C2SMsgType", C2SMsgType_name, C2SMsgType_value)
	proto.RegisterType((*C2SRoomInfo)(nil), "msg.C2SRoomInfo")
	proto.RegisterType((*C2SStartMatch)(nil), "msg.C2SStartMatch")
	proto.RegisterType((*C2SSetCards)(nil), "msg.C2SSetCards")
	proto.RegisterType((*C2SUserSelectCards)(nil), "msg.C2SUserSelectCards")
}

func init() { proto.RegisterFile("game_c2s.proto", fileDescriptor_335bee2329d289c2) }

var fileDescriptor_335bee2329d289c2 = []byte{
	// 281 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x51, 0x4b, 0xf3, 0x30,
	0x14, 0x86, 0xbf, 0xae, 0xdf, 0xc6, 0x76, 0xe6, 0xb4, 0x06, 0x95, 0x21, 0x82, 0x73, 0x57, 0x63,
	0x17, 0x5e, 0xd4, 0x5f, 0x50, 0x62, 0xc5, 0x81, 0xb5, 0x90, 0xd3, 0x5d, 0x97, 0xd8, 0xc6, 0x2e,
	0xd0, 0x9a, 0xd2, 0x64, 0x17, 0xfe, 0x16, 0xff, 0xac, 0x34, 0xad, 0x9b, 0x17, 0xde, 0xe5, 0x7d,
	0x9e, 0xf3, 0x86, 0xc3, 0x81, 0xd3, 0x82, 0x57, 0x22, 0xcd, 0x7c, 0x7d, 0x5f, 0x37, 0xca, 0x28,
	0xe2, 0x56, 0xba, 0x58, 0xde, 0xc2, 0x94, 0xfa, 0xc8, 0x94, 0xaa, 0x36, 0x1f, 0xef, 0x8a, 0x78,
	0xe0, 0xee, 0x65, 0x3e, 0x77, 0x16, 0xce, 0xca, 0x65, 0xed, 0x73, 0x79, 0x07, 0x33, 0xea, 0x23,
	0x1a, 0xde, 0x98, 0x88, 0x9b, 0x6c, 0xf7, 0xc7, 0xc8, 0x97, 0x63, 0x3f, 0x41, 0x61, 0x28, 0x6f,
	0x72, 0x4d, 0x6e, 0x60, 0xb2, 0x13, 0x3c, 0xb7, 0xc1, 0xce, 0x9d, 0xb0, 0x23, 0x20, 0xd7, 0x30,
	0xae, 0x64, 0x2f, 0x07, 0x56, 0x1e, 0x72, 0xdb, 0x34, 0x5c, 0x96, 0x9d, 0x74, 0xbb, 0xe6, 0x01,
	0x90, 0x2b, 0x18, 0x49, 0x1d, 0xec, 0x8d, 0x9a, 0xff, 0x5f, 0x38, 0xab, 0x31, 0xeb, 0x53, 0xdb,
	0x92, 0x1a, 0x6b, 0x91, 0x49, 0x5e, 0xce, 0x87, 0x56, 0x1d, 0xc1, 0x72, 0x0d, 0x84, 0xfa, 0xb8,
	0xd5, 0xa2, 0x41, 0x51, 0x8a, 0xac, 0xdf, 0xf1, 0x02, 0x86, 0xd9, 0xaf, 0xfd, 0xba, 0xb0, 0x66,
	0x00, 0xd4, 0xc7, 0x48, 0x17, 0xc9, 0x67, 0x2d, 0xc8, 0x0c, 0x26, 0x2c, 0x8e, 0xa3, 0x74, 0xf3,
	0xfa, 0x14, 0x7b, 0xff, 0xc8, 0x19, 0x4c, 0x31, 0x09, 0x58, 0x92, 0x46, 0x41, 0x42, 0x9f, 0x3d,
	0xa7, 0xf5, 0x18, 0x26, 0x29, 0x0d, 0xd8, 0x23, 0x7a, 0x03, 0x72, 0x09, 0xe7, 0x5b, 0x0c, 0x59,
	0x8a, 0xe1, 0x4b, 0x48, 0x7f, 0xb0, 0xfb, 0x36, 0xb2, 0xd7, 0x7e, 0xf8, 0x0e, 0x00, 0x00, 0xff,
	0xff, 0xe4, 0x78, 0xcf, 0xcb, 0x7f, 0x01, 0x00, 0x00,
}
