// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.12
// source: ubikom.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Protocol int32

const (
	Protocol_PL_UNKNOWN Protocol = 0
	Protocol_PL_DMS     Protocol = 1
)

// Enum value maps for Protocol.
var (
	Protocol_name = map[int32]string{
		0: "PL_UNKNOWN",
		1: "PL_DMS",
	}
	Protocol_value = map[string]int32{
		"PL_UNKNOWN": 0,
		"PL_DMS":     1,
	}
)

func (x Protocol) Enum() *Protocol {
	p := new(Protocol)
	*p = x
	return p
}

func (x Protocol) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Protocol) Descriptor() protoreflect.EnumDescriptor {
	return file_ubikom_proto_enumTypes[0].Descriptor()
}

func (Protocol) Type() protoreflect.EnumType {
	return &file_ubikom_proto_enumTypes[0]
}

func (x Protocol) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Protocol.Descriptor instead.
func (Protocol) EnumDescriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{0}
}

type ContentWithPOW struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content []byte `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	Pow     []byte `protobuf:"bytes,2,opt,name=pow,proto3" json:"pow,omitempty"`
}

func (x *ContentWithPOW) Reset() {
	*x = ContentWithPOW{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContentWithPOW) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContentWithPOW) ProtoMessage() {}

func (x *ContentWithPOW) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContentWithPOW.ProtoReflect.Descriptor instead.
func (*ContentWithPOW) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{0}
}

func (x *ContentWithPOW) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *ContentWithPOW) GetPow() []byte {
	if x != nil {
		return x.Pow
	}
	return nil
}

type Signature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	R []byte `protobuf:"bytes,1,opt,name=r,proto3" json:"r,omitempty"`
	S []byte `protobuf:"bytes,2,opt,name=s,proto3" json:"s,omitempty"`
}

func (x *Signature) Reset() {
	*x = Signature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signature) ProtoMessage() {}

func (x *Signature) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signature.ProtoReflect.Descriptor instead.
func (*Signature) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{1}
}

func (x *Signature) GetR() []byte {
	if x != nil {
		return x.R
	}
	return nil
}

func (x *Signature) GetS() []byte {
	if x != nil {
		return x.S
	}
	return nil
}

type Signed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content   []byte     `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	Signature *Signature `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	Key       []byte     `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *Signed) Reset() {
	*x = Signed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signed) ProtoMessage() {}

func (x *Signed) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signed.ProtoReflect.Descriptor instead.
func (*Signed) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{2}
}

func (x *Signed) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *Signed) GetSignature() *Signature {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *Signed) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

type SignedWithPow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content   []byte     `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	Signature *Signature `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	Key       []byte     `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"` // Public key used to sign the request.
	Pow       []byte     `protobuf:"bytes,4,opt,name=pow,proto3" json:"pow,omitempty"`
}

func (x *SignedWithPow) Reset() {
	*x = SignedWithPow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignedWithPow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedWithPow) ProtoMessage() {}

func (x *SignedWithPow) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedWithPow.ProtoReflect.Descriptor instead.
func (*SignedWithPow) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{3}
}

func (x *SignedWithPow) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *SignedWithPow) GetSignature() *Signature {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *SignedWithPow) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *SignedWithPow) GetPow() []byte {
	if x != nil {
		return x.Pow
	}
	return nil
}

type LookupKeyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *LookupKeyRequest) Reset() {
	*x = LookupKeyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupKeyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupKeyRequest) ProtoMessage() {}

func (x *LookupKeyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupKeyRequest.ProtoReflect.Descriptor instead.
func (*LookupKeyRequest) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{4}
}

func (x *LookupKeyRequest) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

type LookupKeyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RegistrationTimestamp int64    `protobuf:"varint,1,opt,name=registration_timestamp,json=registrationTimestamp,proto3" json:"registration_timestamp,omitempty"`
	Disabled              bool     `protobuf:"varint,2,opt,name=disabled,proto3" json:"disabled,omitempty"`
	DisabledTimestamp     int64    `protobuf:"varint,3,opt,name=disabled_timestamp,json=disabledTimestamp,proto3" json:"disabled_timestamp,omitempty"`
	DisabledBy            []byte   `protobuf:"bytes,4,opt,name=disabled_by,json=disabledBy,proto3" json:"disabled_by,omitempty"`
	ParentKey             [][]byte `protobuf:"bytes,5,rep,name=parent_key,json=parentKey,proto3" json:"parent_key,omitempty"`
}

func (x *LookupKeyResponse) Reset() {
	*x = LookupKeyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupKeyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupKeyResponse) ProtoMessage() {}

func (x *LookupKeyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupKeyResponse.ProtoReflect.Descriptor instead.
func (*LookupKeyResponse) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{5}
}

func (x *LookupKeyResponse) GetRegistrationTimestamp() int64 {
	if x != nil {
		return x.RegistrationTimestamp
	}
	return 0
}

func (x *LookupKeyResponse) GetDisabled() bool {
	if x != nil {
		return x.Disabled
	}
	return false
}

func (x *LookupKeyResponse) GetDisabledTimestamp() int64 {
	if x != nil {
		return x.DisabledTimestamp
	}
	return 0
}

func (x *LookupKeyResponse) GetDisabledBy() []byte {
	if x != nil {
		return x.DisabledBy
	}
	return nil
}

func (x *LookupKeyResponse) GetParentKey() [][]byte {
	if x != nil {
		return x.ParentKey
	}
	return nil
}

type LookupNameRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *LookupNameRequest) Reset() {
	*x = LookupNameRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupNameRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupNameRequest) ProtoMessage() {}

func (x *LookupNameRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupNameRequest.ProtoReflect.Descriptor instead.
func (*LookupNameRequest) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{6}
}

func (x *LookupNameRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type LookupNameResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *LookupNameResponse) Reset() {
	*x = LookupNameResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupNameResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupNameResponse) ProtoMessage() {}

func (x *LookupNameResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupNameResponse.ProtoReflect.Descriptor instead.
func (*LookupNameResponse) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{7}
}

func (x *LookupNameResponse) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

type LookupAddressRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Protocol Protocol `protobuf:"varint,2,opt,name=protocol,proto3,enum=Ubikom.Protocol" json:"protocol,omitempty"`
}

func (x *LookupAddressRequest) Reset() {
	*x = LookupAddressRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupAddressRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupAddressRequest) ProtoMessage() {}

func (x *LookupAddressRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupAddressRequest.ProtoReflect.Descriptor instead.
func (*LookupAddressRequest) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{8}
}

func (x *LookupAddressRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *LookupAddressRequest) GetProtocol() Protocol {
	if x != nil {
		return x.Protocol
	}
	return Protocol_PL_UNKNOWN
}

type LookupAddressResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *LookupAddressResponse) Reset() {
	*x = LookupAddressResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupAddressResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupAddressResponse) ProtoMessage() {}

func (x *LookupAddressResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupAddressResponse.ProtoReflect.Descriptor instead.
func (*LookupAddressResponse) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{9}
}

func (x *LookupAddressResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *LookupAddressResponse) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

type DMSMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Sender's address.
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// Receiver's address.
	Receiver  string     `protobuf:"bytes,2,opt,name=receiver,proto3" json:"receiver,omitempty"`
	Content   []byte     `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
	Signature *Signature `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *DMSMessage) Reset() {
	*x = DMSMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DMSMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DMSMessage) ProtoMessage() {}

func (x *DMSMessage) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DMSMessage.ProtoReflect.Descriptor instead.
func (*DMSMessage) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{10}
}

func (x *DMSMessage) GetSender() string {
	if x != nil {
		return x.Sender
	}
	return ""
}

func (x *DMSMessage) GetReceiver() string {
	if x != nil {
		return x.Receiver
	}
	return ""
}

func (x *DMSMessage) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *DMSMessage) GetSignature() *Signature {
	if x != nil {
		return x.Signature
	}
	return nil
}

type SendRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message *DMSMessage `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *SendRequest) Reset() {
	*x = SendRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendRequest) ProtoMessage() {}

func (x *SendRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendRequest.ProtoReflect.Descriptor instead.
func (*SendRequest) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{11}
}

func (x *SendRequest) GetMessage() *DMSMessage {
	if x != nil {
		return x.Message
	}
	return nil
}

type SendResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SendResponse) Reset() {
	*x = SendResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendResponse) ProtoMessage() {}

func (x *SendResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendResponse.ProtoReflect.Descriptor instead.
func (*SendResponse) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{12}
}

type ReceiveRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IdentityProof *Signed `protobuf:"bytes,1,opt,name=identity_proof,json=identityProof,proto3" json:"identity_proof,omitempty"`
}

func (x *ReceiveRequest) Reset() {
	*x = ReceiveRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReceiveRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReceiveRequest) ProtoMessage() {}

func (x *ReceiveRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReceiveRequest.ProtoReflect.Descriptor instead.
func (*ReceiveRequest) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{13}
}

func (x *ReceiveRequest) GetIdentityProof() *Signed {
	if x != nil {
		return x.IdentityProof
	}
	return nil
}

type ReceiveResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message *DMSMessage `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ReceiveResponse) Reset() {
	*x = ReceiveResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ubikom_proto_msgTypes[14]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReceiveResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReceiveResponse) ProtoMessage() {}

func (x *ReceiveResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ubikom_proto_msgTypes[14]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReceiveResponse.ProtoReflect.Descriptor instead.
func (*ReceiveResponse) Descriptor() ([]byte, []int) {
	return file_ubikom_proto_rawDescGZIP(), []int{14}
}

func (x *ReceiveResponse) GetMessage() *DMSMessage {
	if x != nil {
		return x.Message
	}
	return nil
}

var File_ubikom_proto protoreflect.FileDescriptor

var file_ubikom_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x75, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x22, 0x3c, 0x0a, 0x0e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x57, 0x69, 0x74, 0x68, 0x50, 0x4f, 0x57, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x70, 0x6f, 0x77, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x03, 0x70, 0x6f, 0x77, 0x22, 0x27, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x12, 0x0c, 0x0a, 0x01, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x72, 0x12,
	0x0c, 0x0a, 0x01, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x73, 0x22, 0x65, 0x0a,
	0x06, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x12, 0x2f, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x53, 0x69,
	0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x22, 0x7e, 0x0a, 0x0d, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x57, 0x69,
	0x74, 0x68, 0x50, 0x6f, 0x77, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12,
	0x2f, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x11, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x53, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x70, 0x6f, 0x77, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x03, 0x70, 0x6f, 0x77, 0x22, 0x24, 0x0a, 0x10, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x4b, 0x65,
	0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0xd5, 0x01, 0x0a, 0x11, 0x4c,
	0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x35, 0x0a, 0x16, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x15, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x69, 0x73, 0x61, 0x62,
	0x6c, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x64, 0x69, 0x73, 0x61, 0x62,
	0x6c, 0x65, 0x64, 0x12, 0x2d, 0x0a, 0x12, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x5f,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x11, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x12, 0x1f, 0x0a, 0x0b, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x5f, 0x62,
	0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65,
	0x64, 0x42, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x6b, 0x65,
	0x79, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x4b,
	0x65, 0x79, 0x22, 0x27, 0x0a, 0x11, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x4e, 0x61, 0x6d, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x26, 0x0a, 0x12, 0x4c,
	0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x22, 0x58, 0x0a, 0x14, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x2c, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x10, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x22, 0x4b, 0x0a,
	0x15, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x8b, 0x01, 0x0a, 0x0a, 0x44,
	0x4d, 0x53, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e,
	0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65,
	0x72, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x12, 0x18, 0x0a,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x2f, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x55, 0x62, 0x69,
	0x6b, 0x6f, 0x6d, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x09, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x22, 0x3b, 0x0a, 0x0b, 0x53, 0x65, 0x6e, 0x64,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f,
	0x6d, 0x2e, 0x44, 0x4d, 0x53, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x0e, 0x0a, 0x0c, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x47, 0x0a, 0x0e, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x35, 0x0a, 0x0e, 0x69, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x5f, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0e, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x52,
	0x0d, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x22, 0x3f,
	0x0a, 0x0f, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x2c, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x44, 0x4d, 0x53, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2a,
	0x26, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x0e, 0x0a, 0x0a, 0x50,
	0x4c, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x50,
	0x4c, 0x5f, 0x44, 0x4d, 0x53, 0x10, 0x01, 0x32, 0xe4, 0x01, 0x0a, 0x0d, 0x4c, 0x6f, 0x6f, 0x6b,
	0x75, 0x70, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x40, 0x0a, 0x09, 0x4c, 0x6f, 0x6f,
	0x6b, 0x75, 0x70, 0x4b, 0x65, 0x79, 0x12, 0x18, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e,
	0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x19, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70,
	0x4b, 0x65, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x43, 0x0a, 0x0a, 0x4c,
	0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x2e, 0x55, 0x62, 0x69, 0x6b,
	0x6f, 0x6d, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x4c, 0x6f,
	0x6f, 0x6b, 0x75, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x4c, 0x0a, 0x0d, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x12, 0x1c, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75,
	0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1d, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x7f,
	0x0a, 0x0e, 0x44, 0x4d, 0x53, 0x44, 0x75, 0x6d, 0x70, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x31, 0x0a, 0x04, 0x53, 0x65, 0x6e, 0x64, 0x12, 0x13, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f,
	0x6d, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e,
	0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x3a, 0x0a, 0x07, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x12, 0x16,
	0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x55, 0x62, 0x69, 0x6b, 0x6f, 0x6d, 0x2e,
	0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42,
	0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ubikom_proto_rawDescOnce sync.Once
	file_ubikom_proto_rawDescData = file_ubikom_proto_rawDesc
)

func file_ubikom_proto_rawDescGZIP() []byte {
	file_ubikom_proto_rawDescOnce.Do(func() {
		file_ubikom_proto_rawDescData = protoimpl.X.CompressGZIP(file_ubikom_proto_rawDescData)
	})
	return file_ubikom_proto_rawDescData
}

var file_ubikom_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_ubikom_proto_msgTypes = make([]protoimpl.MessageInfo, 15)
var file_ubikom_proto_goTypes = []interface{}{
	(Protocol)(0),                 // 0: Ubikom.Protocol
	(*ContentWithPOW)(nil),        // 1: Ubikom.ContentWithPOW
	(*Signature)(nil),             // 2: Ubikom.Signature
	(*Signed)(nil),                // 3: Ubikom.Signed
	(*SignedWithPow)(nil),         // 4: Ubikom.SignedWithPow
	(*LookupKeyRequest)(nil),      // 5: Ubikom.LookupKeyRequest
	(*LookupKeyResponse)(nil),     // 6: Ubikom.LookupKeyResponse
	(*LookupNameRequest)(nil),     // 7: Ubikom.LookupNameRequest
	(*LookupNameResponse)(nil),    // 8: Ubikom.LookupNameResponse
	(*LookupAddressRequest)(nil),  // 9: Ubikom.LookupAddressRequest
	(*LookupAddressResponse)(nil), // 10: Ubikom.LookupAddressResponse
	(*DMSMessage)(nil),            // 11: Ubikom.DMSMessage
	(*SendRequest)(nil),           // 12: Ubikom.SendRequest
	(*SendResponse)(nil),          // 13: Ubikom.SendResponse
	(*ReceiveRequest)(nil),        // 14: Ubikom.ReceiveRequest
	(*ReceiveResponse)(nil),       // 15: Ubikom.ReceiveResponse
}
var file_ubikom_proto_depIdxs = []int32{
	2,  // 0: Ubikom.Signed.signature:type_name -> Ubikom.Signature
	2,  // 1: Ubikom.SignedWithPow.signature:type_name -> Ubikom.Signature
	0,  // 2: Ubikom.LookupAddressRequest.protocol:type_name -> Ubikom.Protocol
	2,  // 3: Ubikom.DMSMessage.signature:type_name -> Ubikom.Signature
	11, // 4: Ubikom.SendRequest.message:type_name -> Ubikom.DMSMessage
	3,  // 5: Ubikom.ReceiveRequest.identity_proof:type_name -> Ubikom.Signed
	11, // 6: Ubikom.ReceiveResponse.message:type_name -> Ubikom.DMSMessage
	5,  // 7: Ubikom.LookupService.LookupKey:input_type -> Ubikom.LookupKeyRequest
	7,  // 8: Ubikom.LookupService.LookupName:input_type -> Ubikom.LookupNameRequest
	9,  // 9: Ubikom.LookupService.LookupAddress:input_type -> Ubikom.LookupAddressRequest
	12, // 10: Ubikom.DMSDumpService.Send:input_type -> Ubikom.SendRequest
	14, // 11: Ubikom.DMSDumpService.Receive:input_type -> Ubikom.ReceiveRequest
	6,  // 12: Ubikom.LookupService.LookupKey:output_type -> Ubikom.LookupKeyResponse
	8,  // 13: Ubikom.LookupService.LookupName:output_type -> Ubikom.LookupNameResponse
	10, // 14: Ubikom.LookupService.LookupAddress:output_type -> Ubikom.LookupAddressResponse
	13, // 15: Ubikom.DMSDumpService.Send:output_type -> Ubikom.SendResponse
	15, // 16: Ubikom.DMSDumpService.Receive:output_type -> Ubikom.ReceiveResponse
	12, // [12:17] is the sub-list for method output_type
	7,  // [7:12] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_ubikom_proto_init() }
func file_ubikom_proto_init() {
	if File_ubikom_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ubikom_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContentWithPOW); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signature); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signed); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignedWithPow); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupKeyRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupKeyResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupNameRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupNameResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupAddressRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupAddressResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DMSMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReceiveRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ubikom_proto_msgTypes[14].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReceiveResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ubikom_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   15,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_ubikom_proto_goTypes,
		DependencyIndexes: file_ubikom_proto_depIdxs,
		EnumInfos:         file_ubikom_proto_enumTypes,
		MessageInfos:      file_ubikom_proto_msgTypes,
	}.Build()
	File_ubikom_proto = out.File
	file_ubikom_proto_rawDesc = nil
	file_ubikom_proto_goTypes = nil
	file_ubikom_proto_depIdxs = nil
}
