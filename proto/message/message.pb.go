// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message/message.proto

package message

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type UnitMessage_MessageFlag int32

const (
	UnitMessage_COMMON     UnitMessage_MessageFlag = 0
	UnitMessage_SYNC       UnitMessage_MessageFlag = 1
	UnitMessage_HEART_BEAT UnitMessage_MessageFlag = 2
	UnitMessage_REAL_TIME  UnitMessage_MessageFlag = 3
	UnitMessage_CLOSE      UnitMessage_MessageFlag = 11
)

var UnitMessage_MessageFlag_name = map[int32]string{
	0:  "COMMON",
	1:  "SYNC",
	2:  "HEART_BEAT",
	3:  "REAL_TIME",
	11: "CLOSE",
}

var UnitMessage_MessageFlag_value = map[string]int32{
	"COMMON":     0,
	"SYNC":       1,
	"HEART_BEAT": 2,
	"REAL_TIME":  3,
	"CLOSE":      11,
}

func (x UnitMessage_MessageFlag) String() string {
	return proto.EnumName(UnitMessage_MessageFlag_name, int32(x))
}

func (UnitMessage_MessageFlag) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_ebceca9e8703e37f, []int{0, 0}
}

type UnitMessage_MessageType int32

const (
	UnitMessage_BroadCast    UnitMessage_MessageType = 0
	UnitMessage_PointToPoint UnitMessage_MessageType = 1
	UnitMessage_Group        UnitMessage_MessageType = 2
)

var UnitMessage_MessageType_name = map[int32]string{
	0: "BroadCast",
	1: "PointToPoint",
	2: "Group",
}

var UnitMessage_MessageType_value = map[string]int32{
	"BroadCast":    0,
	"PointToPoint": 1,
	"Group":        2,
}

func (x UnitMessage_MessageType) String() string {
	return proto.EnumName(UnitMessage_MessageType_name, int32(x))
}

func (UnitMessage_MessageType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_ebceca9e8703e37f, []int{0, 1}
}

type UnitMessage struct {
	ChannelId            string                  `protobuf:"bytes,1,opt,name=ChannelId,proto3" json:"ChannelId,omitempty"`
	SrcEndPointId        string                  `protobuf:"bytes,2,opt,name=SrcEndPointId,proto3" json:"SrcEndPointId,omitempty"`
	DstEndPointId        []string                `protobuf:"bytes,3,rep,name=DstEndPointId,proto3" json:"DstEndPointId,omitempty"`
	Flag                 UnitMessage_MessageFlag `protobuf:"varint,4,opt,name=Flag,proto3,enum=redresseur.message_bus.UnitMessage_MessageFlag" json:"Flag,omitempty"`
	Type                 UnitMessage_MessageType `protobuf:"varint,5,opt,name=Type,proto3,enum=redresseur.message_bus.UnitMessage_MessageType" json:"Type,omitempty"`
	Seq                  uint32                  `protobuf:"varint,6,opt,name=Seq,proto3" json:"Seq,omitempty"`
	Ack                  uint32                  `protobuf:"varint,7,opt,name=Ack,proto3" json:"Ack,omitempty"`
	Payload              []byte                  `protobuf:"bytes,8,opt,name=Payload,proto3" json:"Payload,omitempty"`
	Metadata             []byte                  `protobuf:"bytes,9,opt,name=Metadata,proto3" json:"Metadata,omitempty"`
	Timestamp            *timestamp.Timestamp    `protobuf:"bytes,10,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *UnitMessage) Reset()         { *m = UnitMessage{} }
func (m *UnitMessage) String() string { return proto.CompactTextString(m) }
func (*UnitMessage) ProtoMessage()    {}
func (*UnitMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebceca9e8703e37f, []int{0}
}

func (m *UnitMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UnitMessage.Unmarshal(m, b)
}
func (m *UnitMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UnitMessage.Marshal(b, m, deterministic)
}
func (m *UnitMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UnitMessage.Merge(m, src)
}
func (m *UnitMessage) XXX_Size() int {
	return xxx_messageInfo_UnitMessage.Size(m)
}
func (m *UnitMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_UnitMessage.DiscardUnknown(m)
}

var xxx_messageInfo_UnitMessage proto.InternalMessageInfo

func (m *UnitMessage) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

func (m *UnitMessage) GetSrcEndPointId() string {
	if m != nil {
		return m.SrcEndPointId
	}
	return ""
}

func (m *UnitMessage) GetDstEndPointId() []string {
	if m != nil {
		return m.DstEndPointId
	}
	return nil
}

func (m *UnitMessage) GetFlag() UnitMessage_MessageFlag {
	if m != nil {
		return m.Flag
	}
	return UnitMessage_COMMON
}

func (m *UnitMessage) GetType() UnitMessage_MessageType {
	if m != nil {
		return m.Type
	}
	return UnitMessage_BroadCast
}

func (m *UnitMessage) GetSeq() uint32 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *UnitMessage) GetAck() uint32 {
	if m != nil {
		return m.Ack
	}
	return 0
}

func (m *UnitMessage) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *UnitMessage) GetMetadata() []byte {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *UnitMessage) GetTimestamp() *timestamp.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func init() {
	proto.RegisterEnum("redresseur.message_bus.UnitMessage_MessageFlag", UnitMessage_MessageFlag_name, UnitMessage_MessageFlag_value)
	proto.RegisterEnum("redresseur.message_bus.UnitMessage_MessageType", UnitMessage_MessageType_name, UnitMessage_MessageType_value)
	proto.RegisterType((*UnitMessage)(nil), "redresseur.message_bus.UnitMessage")
}

func init() { proto.RegisterFile("message/message.proto", fileDescriptor_ebceca9e8703e37f) }

var fileDescriptor_ebceca9e8703e37f = []byte{
	// 408 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x91, 0xcf, 0x6f, 0xda, 0x30,
	0x14, 0xc7, 0x6b, 0xa0, 0x94, 0x3c, 0x4a, 0x65, 0x59, 0xda, 0xe4, 0xa1, 0x1d, 0x22, 0xb4, 0x43,
	0x4e, 0x8e, 0xd6, 0x5d, 0xb6, 0x63, 0xc8, 0xb2, 0x0d, 0xa9, 0x29, 0x95, 0xc9, 0x0e, 0xdb, 0x05,
	0x19, 0xe2, 0xa5, 0xa8, 0x10, 0xb3, 0xd8, 0x39, 0xf4, 0xaf, 0xd9, 0xbf, 0x3a, 0xd9, 0x40, 0x48,
	0xa5, 0x5d, 0x76, 0x7a, 0x7e, 0x5f, 0x7f, 0xbe, 0x4f, 0xef, 0x07, 0xbc, 0xda, 0x49, 0xad, 0x45,
	0x21, 0xc3, 0x63, 0x64, 0xfb, 0x4a, 0x19, 0x45, 0x5e, 0x57, 0x32, 0xaf, 0xa4, 0xd6, 0xb2, 0xae,
	0xd8, 0xf1, 0x67, 0xb9, 0xaa, 0xf5, 0xf8, 0x8d, 0xd9, 0xec, 0xa4, 0x36, 0x62, 0xb7, 0x0f, 0x9b,
	0xd7, 0xc1, 0x32, 0xf9, 0xd3, 0x83, 0xe1, 0xf7, 0x72, 0x63, 0xd2, 0x03, 0x4e, 0xde, 0x82, 0x17,
	0x3f, 0x8a, 0xb2, 0x94, 0xdb, 0x59, 0x4e, 0x91, 0x8f, 0x02, 0x8f, 0x9f, 0x05, 0xf2, 0x0e, 0x46,
	0x8b, 0x6a, 0x9d, 0x94, 0xf9, 0x83, 0xda, 0x94, 0x66, 0x96, 0xd3, 0x8e, 0x23, 0x5e, 0x8a, 0x96,
	0xfa, 0xac, 0x4d, 0x8b, 0xea, 0xfa, 0x5d, 0x4b, 0xbd, 0x10, 0x49, 0x0c, 0xbd, 0x2f, 0x5b, 0x51,
	0xd0, 0x9e, 0x8f, 0x82, 0x9b, 0xdb, 0x90, 0xfd, 0xbb, 0x77, 0xd6, 0x6a, 0x8e, 0x1d, 0xa3, 0xb5,
	0x71, 0x67, 0xb6, 0x45, 0xb2, 0xe7, 0xbd, 0xa4, 0x97, 0xff, 0x5d, 0xc4, 0xda, 0xb8, 0x33, 0x13,
	0x0c, 0xdd, 0x85, 0xfc, 0x4d, 0xfb, 0x3e, 0x0a, 0x46, 0xdc, 0x3e, 0xad, 0x12, 0xad, 0x9f, 0xe8,
	0xd5, 0x41, 0x89, 0xd6, 0x4f, 0x84, 0xc2, 0xd5, 0x83, 0x78, 0xde, 0x2a, 0x91, 0xd3, 0x81, 0x8f,
	0x82, 0x6b, 0x7e, 0x4a, 0xc9, 0x18, 0x06, 0xa9, 0x34, 0x22, 0x17, 0x46, 0x50, 0xcf, 0x7d, 0x35,
	0x39, 0xf9, 0x08, 0x5e, 0xb3, 0x70, 0x0a, 0x3e, 0x0a, 0x86, 0xb7, 0x63, 0x56, 0x28, 0x55, 0x6c,
	0x8f, 0x27, 0x5b, 0xd5, 0xbf, 0x58, 0x76, 0x22, 0xf8, 0x19, 0x9e, 0xa4, 0x30, 0x6c, 0x4d, 0x4b,
	0x00, 0xfa, 0xf1, 0x3c, 0x4d, 0xe7, 0xf7, 0xf8, 0x82, 0x0c, 0xa0, 0xb7, 0xf8, 0x71, 0x1f, 0x63,
	0x44, 0x6e, 0x00, 0xbe, 0x25, 0x11, 0xcf, 0x96, 0xd3, 0x24, 0xca, 0x70, 0x87, 0x8c, 0xc0, 0xe3,
	0x49, 0x74, 0xb7, 0xcc, 0x66, 0x69, 0x82, 0xbb, 0xc4, 0x83, 0xcb, 0xf8, 0x6e, 0xbe, 0x48, 0xf0,
	0x70, 0xf2, 0xa9, 0x29, 0xe7, 0x26, 0x1e, 0x81, 0x37, 0xad, 0x94, 0xc8, 0x63, 0xa1, 0x0d, 0xbe,
	0x20, 0x18, 0xae, 0xdd, 0x55, 0x32, 0xe5, 0x02, 0x46, 0xd6, 0xfa, 0xb5, 0x52, 0xf5, 0x1e, 0x77,
	0xa6, 0xef, 0x7f, 0x86, 0xc5, 0xc6, 0x3c, 0xd6, 0x2b, 0xb6, 0x56, 0xbb, 0xf0, 0xbc, 0xe0, 0xb0,
	0xb5, 0xe0, 0xd0, 0x0d, 0x73, 0x52, 0x56, 0x7d, 0x97, 0x7e, 0xf8, 0x1b, 0x00, 0x00, 0xff, 0xff,
	0x3a, 0xe1, 0xbc, 0xdb, 0xa7, 0x02, 0x00, 0x00,
}
