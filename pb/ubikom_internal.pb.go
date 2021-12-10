// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ubikom_internal.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type KeyRecord struct {
	// Timestamp when the key was registered.
	RegistrationTimestamp int64 `protobuf:"varint,1,opt,name=registration_timestamp,json=registrationTimestamp" json:"registration_timestamp,omitempty"`
	// Controls if the key is disabled. Once a key is disabled, it is dead forever.
	Disabled          bool   `protobuf:"varint,2,opt,name=disabled" json:"disabled,omitempty"`
	DisabledTimestamp int64  `protobuf:"varint,3,opt,name=disabled_timestamp,json=disabledTimestamp" json:"disabled_timestamp,omitempty"`
	DisabledBy        []byte `protobuf:"bytes,4,opt,name=disabled_by,json=disabledBy,proto3" json:"disabled_by,omitempty"`
	// Parent key (it can do everything this key can do).
	ParentKey [][]byte `protobuf:"bytes,5,rep,name=parent_key,json=parentKey,proto3" json:"parent_key,omitempty"`
}

func (m *KeyRecord) Reset()                    { *m = KeyRecord{} }
func (m *KeyRecord) String() string            { return proto.CompactTextString(m) }
func (*KeyRecord) ProtoMessage()               {}
func (*KeyRecord) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *KeyRecord) GetRegistrationTimestamp() int64 {
	if m != nil {
		return m.RegistrationTimestamp
	}
	return 0
}

func (m *KeyRecord) GetDisabled() bool {
	if m != nil {
		return m.Disabled
	}
	return false
}

func (m *KeyRecord) GetDisabledTimestamp() int64 {
	if m != nil {
		return m.DisabledTimestamp
	}
	return 0
}

func (m *KeyRecord) GetDisabledBy() []byte {
	if m != nil {
		return m.DisabledBy
	}
	return nil
}

func (m *KeyRecord) GetParentKey() [][]byte {
	if m != nil {
		return m.ParentKey
	}
	return nil
}

type ExportKeyRecord struct {
	Key                   []byte   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	RegistrationTimestamp int64    `protobuf:"varint,2,opt,name=registration_timestamp,json=registrationTimestamp" json:"registration_timestamp,omitempty"`
	Disabled              bool     `protobuf:"varint,3,opt,name=disabled" json:"disabled,omitempty"`
	DisabledTimestamp     int64    `protobuf:"varint,4,opt,name=disabled_timestamp,json=disabledTimestamp" json:"disabled_timestamp,omitempty"`
	DisabledBy            []byte   `protobuf:"bytes,5,opt,name=disabled_by,json=disabledBy,proto3" json:"disabled_by,omitempty"`
	ParentKey             [][]byte `protobuf:"bytes,6,rep,name=parent_key,json=parentKey,proto3" json:"parent_key,omitempty"`
}

func (m *ExportKeyRecord) Reset()                    { *m = ExportKeyRecord{} }
func (m *ExportKeyRecord) String() string            { return proto.CompactTextString(m) }
func (*ExportKeyRecord) ProtoMessage()               {}
func (*ExportKeyRecord) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *ExportKeyRecord) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *ExportKeyRecord) GetRegistrationTimestamp() int64 {
	if m != nil {
		return m.RegistrationTimestamp
	}
	return 0
}

func (m *ExportKeyRecord) GetDisabled() bool {
	if m != nil {
		return m.Disabled
	}
	return false
}

func (m *ExportKeyRecord) GetDisabledTimestamp() int64 {
	if m != nil {
		return m.DisabledTimestamp
	}
	return 0
}

func (m *ExportKeyRecord) GetDisabledBy() []byte {
	if m != nil {
		return m.DisabledBy
	}
	return nil
}

func (m *ExportKeyRecord) GetParentKey() [][]byte {
	if m != nil {
		return m.ParentKey
	}
	return nil
}

type ExportNameRecord struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Key  []byte `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *ExportNameRecord) Reset()                    { *m = ExportNameRecord{} }
func (m *ExportNameRecord) String() string            { return proto.CompactTextString(m) }
func (*ExportNameRecord) ProtoMessage()               {}
func (*ExportNameRecord) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *ExportNameRecord) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ExportNameRecord) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type ExportAddressRecord struct {
	Name     string   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Protocol Protocol `protobuf:"varint,2,opt,name=protocol,enum=Ubikom.Protocol" json:"protocol,omitempty"`
	Address  string   `protobuf:"bytes,3,opt,name=address" json:"address,omitempty"`
}

func (m *ExportAddressRecord) Reset()                    { *m = ExportAddressRecord{} }
func (m *ExportAddressRecord) String() string            { return proto.CompactTextString(m) }
func (*ExportAddressRecord) ProtoMessage()               {}
func (*ExportAddressRecord) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

func (m *ExportAddressRecord) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ExportAddressRecord) GetProtocol() Protocol {
	if m != nil {
		return m.Protocol
	}
	return Protocol_PL_UNKNOWN
}

func (m *ExportAddressRecord) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type DBValue struct {
	Timestamp uint64               `protobuf:"varint,1,opt,name=timestamp" json:"timestamp,omitempty"`
	Payload   *google_protobuf.Any `protobuf:"bytes,2,opt,name=payload" json:"payload,omitempty"`
}

func (m *DBValue) Reset()                    { *m = DBValue{} }
func (m *DBValue) String() string            { return proto.CompactTextString(m) }
func (*DBValue) ProtoMessage()               {}
func (*DBValue) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{4} }

func (m *DBValue) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *DBValue) GetPayload() *google_protobuf.Any {
	if m != nil {
		return m.Payload
	}
	return nil
}

type DBEntry struct {
	Value []*DBValue `protobuf:"bytes,1,rep,name=value" json:"value,omitempty"`
}

func (m *DBEntry) Reset()                    { *m = DBEntry{} }
func (m *DBEntry) String() string            { return proto.CompactTextString(m) }
func (*DBEntry) ProtoMessage()               {}
func (*DBEntry) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{5} }

func (m *DBEntry) GetValue() []*DBValue {
	if m != nil {
		return m.Value
	}
	return nil
}

type ImapInfo struct {
	NextMailboxUid uint32 `protobuf:"varint,1,opt,name=nextMailboxUid" json:"nextMailboxUid,omitempty"`
	NextMessageUid uint32 `protobuf:"varint,2,opt,name=nextMessageUid" json:"nextMessageUid,omitempty"`
}

func (m *ImapInfo) Reset()                    { *m = ImapInfo{} }
func (m *ImapInfo) String() string            { return proto.CompactTextString(m) }
func (*ImapInfo) ProtoMessage()               {}
func (*ImapInfo) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{6} }

func (m *ImapInfo) GetNextMailboxUid() uint32 {
	if m != nil {
		return m.NextMailboxUid
	}
	return 0
}

func (m *ImapInfo) GetNextMessageUid() uint32 {
	if m != nil {
		return m.NextMessageUid
	}
	return 0
}

type ImapMailbox struct {
	Name           string   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Attribute      []string `protobuf:"bytes,2,rep,name=attribute" json:"attribute,omitempty"`
	Uid            uint32   `protobuf:"varint,3,opt,name=uid" json:"uid,omitempty"`
	NextMessageUid uint32   `protobuf:"varint,4,opt,name=nextMessageUid" json:"nextMessageUid,omitempty"`
}

func (m *ImapMailbox) Reset()                    { *m = ImapMailbox{} }
func (m *ImapMailbox) String() string            { return proto.CompactTextString(m) }
func (*ImapMailbox) ProtoMessage()               {}
func (*ImapMailbox) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{7} }

func (m *ImapMailbox) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ImapMailbox) GetAttribute() []string {
	if m != nil {
		return m.Attribute
	}
	return nil
}

func (m *ImapMailbox) GetUid() uint32 {
	if m != nil {
		return m.Uid
	}
	return 0
}

func (m *ImapMailbox) GetNextMessageUid() uint32 {
	if m != nil {
		return m.NextMessageUid
	}
	return 0
}

type ImapMailboxes struct {
	Mailbox        []*ImapMailbox `protobuf:"bytes,1,rep,name=mailbox" json:"mailbox,omitempty"`
	Subscribed     []string       `protobuf:"bytes,2,rep,name=subscribed" json:"subscribed,omitempty"`
	NextMailboxUid uint32         `protobuf:"varint,3,opt,name=nextMailboxUid" json:"nextMailboxUid,omitempty"`
}

func (m *ImapMailboxes) Reset()                    { *m = ImapMailboxes{} }
func (m *ImapMailboxes) String() string            { return proto.CompactTextString(m) }
func (*ImapMailboxes) ProtoMessage()               {}
func (*ImapMailboxes) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{8} }

func (m *ImapMailboxes) GetMailbox() []*ImapMailbox {
	if m != nil {
		return m.Mailbox
	}
	return nil
}

func (m *ImapMailboxes) GetSubscribed() []string {
	if m != nil {
		return m.Subscribed
	}
	return nil
}

func (m *ImapMailboxes) GetNextMailboxUid() uint32 {
	if m != nil {
		return m.NextMailboxUid
	}
	return 0
}

type ImapMessage struct {
	Content           []byte   `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	Flag              []string `protobuf:"bytes,2,rep,name=flag" json:"flag,omitempty"`
	ReceivedTimestamp uint64   `protobuf:"varint,3,opt,name=received_timestamp,json=receivedTimestamp" json:"received_timestamp,omitempty"`
	Size              uint64   `protobuf:"varint,4,opt,name=size" json:"size,omitempty"`
	Uid               uint32   `protobuf:"varint,5,opt,name=uid" json:"uid,omitempty"`
}

func (m *ImapMessage) Reset()                    { *m = ImapMessage{} }
func (m *ImapMessage) String() string            { return proto.CompactTextString(m) }
func (*ImapMessage) ProtoMessage()               {}
func (*ImapMessage) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{9} }

func (m *ImapMessage) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

func (m *ImapMessage) GetFlag() []string {
	if m != nil {
		return m.Flag
	}
	return nil
}

func (m *ImapMessage) GetReceivedTimestamp() uint64 {
	if m != nil {
		return m.ReceivedTimestamp
	}
	return 0
}

func (m *ImapMessage) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *ImapMessage) GetUid() uint32 {
	if m != nil {
		return m.Uid
	}
	return 0
}

func init() {
	proto.RegisterType((*KeyRecord)(nil), "Ubikom.KeyRecord")
	proto.RegisterType((*ExportKeyRecord)(nil), "Ubikom.ExportKeyRecord")
	proto.RegisterType((*ExportNameRecord)(nil), "Ubikom.ExportNameRecord")
	proto.RegisterType((*ExportAddressRecord)(nil), "Ubikom.ExportAddressRecord")
	proto.RegisterType((*DBValue)(nil), "Ubikom.DBValue")
	proto.RegisterType((*DBEntry)(nil), "Ubikom.DBEntry")
	proto.RegisterType((*ImapInfo)(nil), "Ubikom.ImapInfo")
	proto.RegisterType((*ImapMailbox)(nil), "Ubikom.ImapMailbox")
	proto.RegisterType((*ImapMailboxes)(nil), "Ubikom.ImapMailboxes")
	proto.RegisterType((*ImapMessage)(nil), "Ubikom.ImapMessage")
}

func init() { proto.RegisterFile("ubikom_internal.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 601 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0xdd, 0x6e, 0xd4, 0x3c,
	0x10, 0x55, 0x36, 0xd9, 0xee, 0x66, 0xb6, 0xbf, 0xee, 0xd7, 0x4f, 0xa1, 0x2a, 0xb0, 0x8a, 0x04,
	0xda, 0x0b, 0x9a, 0xa2, 0x22, 0x24, 0x24, 0xae, 0x5a, 0xd1, 0x8b, 0xaa, 0x02, 0x21, 0x8b, 0x82,
	0xd4, 0x9b, 0xca, 0xd9, 0x75, 0x57, 0x56, 0x13, 0x3b, 0x38, 0x4e, 0xd5, 0xf0, 0x00, 0x5c, 0xf3,
	0x62, 0x3c, 0x06, 0xef, 0x81, 0x6c, 0xc7, 0xd9, 0x50, 0x56, 0xa5, 0x77, 0x33, 0x67, 0xc6, 0x33,
	0x3e, 0x27, 0xc7, 0x81, 0x9d, 0x2a, 0x65, 0xd7, 0x22, 0xbf, 0x64, 0x5c, 0x51, 0xc9, 0x49, 0x96,
	0x14, 0x52, 0x28, 0x81, 0x56, 0xce, 0x0d, 0xbc, 0xbb, 0x6a, 0xcb, 0x16, 0xdd, 0x7d, 0x34, 0x17,
	0x62, 0x9e, 0xd1, 0x03, 0x93, 0xa5, 0xd5, 0xd5, 0x01, 0xe1, 0xb5, 0x2d, 0xc5, 0x3f, 0x3d, 0x08,
	0xcf, 0x68, 0x8d, 0xe9, 0x54, 0xc8, 0x19, 0x7a, 0x0d, 0xff, 0x4b, 0x3a, 0x67, 0xa5, 0x92, 0x44,
	0x31, 0xc1, 0x2f, 0x15, 0xcb, 0x69, 0xa9, 0x48, 0x5e, 0x44, 0xde, 0xd8, 0x9b, 0xf8, 0x78, 0xa7,
	0x5b, 0xfd, 0xe4, 0x8a, 0x68, 0x17, 0x86, 0x33, 0x56, 0x92, 0x34, 0xa3, 0xb3, 0xa8, 0x37, 0xf6,
	0x26, 0x43, 0xdc, 0xe6, 0x68, 0x1f, 0x90, 0x8b, 0x3b, 0xe3, 0x7c, 0x33, 0x6e, 0xcb, 0x55, 0x16,
	0xa3, 0x9e, 0xc2, 0xa8, 0x6d, 0x4f, 0xeb, 0x28, 0x18, 0x7b, 0x93, 0x55, 0x0c, 0x0e, 0x3a, 0xae,
	0xd1, 0x63, 0x80, 0x82, 0x48, 0xca, 0xd5, 0xe5, 0x35, 0xad, 0xa3, 0xfe, 0xd8, 0x9f, 0xac, 0xe2,
	0xd0, 0x22, 0x67, 0xb4, 0x8e, 0x7f, 0x79, 0xb0, 0x71, 0x72, 0x5b, 0x08, 0xa9, 0x16, 0xac, 0x36,
	0xc1, 0xd7, 0xbd, 0x9e, 0x99, 0xa5, 0xc3, 0x7b, 0x78, 0xf6, 0x1e, 0xca, 0xd3, 0x7f, 0x10, 0xcf,
	0xe0, 0x81, 0x3c, 0xfb, 0xff, 0xe0, 0xb9, 0x72, 0x97, 0xe7, 0x1b, 0xd8, 0xb4, 0x34, 0x3f, 0x90,
	0x9c, 0x36, 0x3c, 0x11, 0x04, 0x9c, 0xe4, 0xd4, 0x10, 0x0d, 0xb1, 0x89, 0x1d, 0xf7, 0x5e, 0xcb,
	0x3d, 0xfe, 0x0a, 0xdb, 0xf6, 0xe4, 0xd1, 0x6c, 0x26, 0x69, 0x59, 0xde, 0x73, 0xf8, 0x05, 0x0c,
	0x8d, 0x4b, 0xa6, 0x22, 0x33, 0x13, 0xd6, 0x0f, 0x37, 0x13, 0x6b, 0xb0, 0xe4, 0x63, 0x83, 0xe3,
	0xb6, 0x03, 0x45, 0x30, 0x20, 0x76, 0xa4, 0x11, 0x27, 0xc4, 0x2e, 0x8d, 0xbf, 0xc0, 0xe0, 0xdd,
	0xf1, 0x67, 0x92, 0x55, 0x14, 0xed, 0x41, 0xf8, 0xa7, 0xa9, 0x02, 0xbc, 0x00, 0x50, 0x02, 0x83,
	0x82, 0xd4, 0x99, 0x20, 0xd6, 0x47, 0xa3, 0xc3, 0xff, 0x12, 0x6b, 0xdd, 0xc4, 0x59, 0x37, 0x39,
	0xe2, 0x35, 0x76, 0x4d, 0xf1, 0x4b, 0x3d, 0xf8, 0x84, 0x2b, 0x59, 0xa3, 0x67, 0xd0, 0xbf, 0xd1,
	0x1b, 0x22, 0x6f, 0xec, 0x4f, 0x46, 0x87, 0x1b, 0xee, 0xa2, 0xcd, 0x62, 0x6c, 0xab, 0xf1, 0x05,
	0x0c, 0x4f, 0x73, 0x52, 0x9c, 0xf2, 0x2b, 0x81, 0x9e, 0xc3, 0x3a, 0xa7, 0xb7, 0xea, 0x3d, 0x61,
	0x59, 0x2a, 0x6e, 0xcf, 0xd9, 0xcc, 0x5c, 0x68, 0x0d, 0xdf, 0x41, 0xdb, 0x3e, 0x5a, 0x96, 0x64,
	0x4e, 0x75, 0x5f, 0xaf, 0xd3, 0xd7, 0xa2, 0x71, 0x0d, 0x23, 0x3d, 0xbb, 0x39, 0xb9, 0x54, 0xd1,
	0x3d, 0x08, 0x89, 0x52, 0x92, 0xa5, 0x95, 0xa2, 0x51, 0x6f, 0xec, 0x4f, 0x42, 0xbc, 0x00, 0xf4,
	0xc7, 0xaa, 0x98, 0xb5, 0xd6, 0x1a, 0xd6, 0xe1, 0x92, 0xd5, 0xc1, 0xd2, 0xd5, 0xdf, 0x3d, 0x58,
	0xeb, 0xec, 0xa6, 0x25, 0xda, 0x87, 0x41, 0x6e, 0x93, 0x46, 0x91, 0x6d, 0xa7, 0x48, 0xa7, 0x0f,
	0xbb, 0x1e, 0xf4, 0x04, 0xa0, 0xac, 0xd2, 0x72, 0x2a, 0x59, 0x6a, 0x1e, 0xb1, 0xbe, 0x59, 0x07,
	0x59, 0xa2, 0x95, 0xbf, 0x4c, 0xab, 0xf8, 0x87, 0xd7, 0x88, 0x60, 0xef, 0xa6, 0x4d, 0x31, 0x15,
	0x5c, 0x51, 0xae, 0x9a, 0xf7, 0xe7, 0x52, 0x2d, 0xcf, 0x55, 0x46, 0xe6, 0xcd, 0x2e, 0x13, 0xeb,
	0x47, 0x24, 0xe9, 0x94, 0xb2, 0x9b, 0xbf, 0x7e, 0x16, 0x01, 0xde, 0x72, 0x95, 0xc5, 0x23, 0x42,
	0x10, 0x94, 0xec, 0x1b, 0x35, 0x9a, 0x04, 0xd8, 0xc4, 0x4e, 0xc3, 0x7e, 0xab, 0xe1, 0xf1, 0xe0,
	0xa2, 0x9f, 0x1c, 0xbc, 0x2d, 0xd2, 0x74, 0xc5, 0x98, 0xe8, 0xd5, 0xef, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xbb, 0x69, 0x64, 0x88, 0x3c, 0x05, 0x00, 0x00,
}
