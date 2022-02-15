// The definition of the communication service between mocking service and manager

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.12.4
// source: management.proto

package management

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

// This client request message provides the node name and fleet ID.
type ClientParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Client node name
	NodeName string `protobuf:"bytes,1,opt,name=nodeName,proto3" json:"nodeName,omitempty"`
	FleetID  string `protobuf:"bytes,2,opt,name=fleetID,proto3" json:"fleetID,omitempty"`
}

func (x *ClientParams) Reset() {
	*x = ClientParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_management_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientParams) ProtoMessage() {}

func (x *ClientParams) ProtoReflect() protoreflect.Message {
	mi := &file_management_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientParams.ProtoReflect.Descriptor instead.
func (*ClientParams) Descriptor() ([]byte, []int) {
	return file_management_proto_rawDescGZIP(), []int{0}
}

func (x *ClientParams) GetNodeName() string {
	if x != nil {
		return x.NodeName
	}
	return ""
}

func (x *ClientParams) GetFleetID() string {
	if x != nil {
		return x.FleetID
	}
	return ""
}

// Snapshot response message provides the snapshot of the configurations.
type Snapshot struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MockConfig *MockConfig `protobuf:"bytes,1,opt,name=mockConfig,proto3" json:"mockConfig,omitempty"`
}

func (x *Snapshot) Reset() {
	*x = Snapshot{}
	if protoimpl.UnsafeEnabled {
		mi := &file_management_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Snapshot) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Snapshot) ProtoMessage() {}

func (x *Snapshot) ProtoReflect() protoreflect.Message {
	mi := &file_management_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Snapshot.ProtoReflect.Descriptor instead.
func (*Snapshot) Descriptor() ([]byte, []int) {
	return file_management_proto_rawDescGZIP(), []int{1}
}

func (x *Snapshot) GetMockConfig() *MockConfig {
	if x != nil {
		return x.MockConfig
	}
	return nil
}

// MockConfig is the mapping of mockID to MockResponse struct
type MockConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MockResponses map[string]*MockResponse `protobuf:"bytes,1,rep,name=MockResponses,proto3" json:"MockResponses,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *MockConfig) Reset() {
	*x = MockConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_management_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MockConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MockConfig) ProtoMessage() {}

func (x *MockConfig) ProtoReflect() protoreflect.Message {
	mi := &file_management_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MockConfig.ProtoReflect.Descriptor instead.
func (*MockConfig) Descriptor() ([]byte, []int) {
	return file_management_proto_rawDescGZIP(), []int{2}
}

func (x *MockConfig) GetMockResponses() map[string]*MockResponse {
	if x != nil {
		return x.MockResponses
	}
	return nil
}

// MockResponse is the mocking.MockResponse struct
type MockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// HTTP Status Code
	StatusCode uint32 `protobuf:"varint,1,opt,name=StatusCode,proto3" json:"StatusCode,omitempty"`
	// Mapping of Media type to Media data
	// application/json -> []byte
	// application/xml -> []byte
	MediaTypeData map[string][]byte `protobuf:"bytes,2,rep,name=MediaTypeData,proto3" json:"MediaTypeData,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *MockResponse) Reset() {
	*x = MockResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_management_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MockResponse) ProtoMessage() {}

func (x *MockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_management_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MockResponse.ProtoReflect.Descriptor instead.
func (*MockResponse) Descriptor() ([]byte, []int) {
	return file_management_proto_rawDescGZIP(), []int{3}
}

func (x *MockResponse) GetStatusCode() uint32 {
	if x != nil {
		return x.StatusCode
	}
	return 0
}

func (x *MockResponse) GetMediaTypeData() map[string][]byte {
	if x != nil {
		return x.MediaTypeData
	}
	return nil
}

var File_management_proto protoreflect.FileDescriptor

var file_management_proto_rawDesc = []byte{
	0x0a, 0x10, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x04, 0x67, 0x72, 0x70, 0x63, 0x22, 0x44, 0x0a, 0x0c, 0x43, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x6e, 0x6f, 0x64, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6e, 0x6f, 0x64, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x49, 0x44, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x49, 0x44, 0x22, 0x3c,
	0x0a, 0x08, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x12, 0x30, 0x0a, 0x0a, 0x6d, 0x6f,
	0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x4d, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x0a, 0x6d, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0xad, 0x01, 0x0a,
	0x0a, 0x4d, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x49, 0x0a, 0x0d, 0x4d,
	0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x23, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x4d, 0x6f, 0x63, 0x6b, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0d, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x1a, 0x54, 0x0a, 0x12, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x28,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e,
	0x67, 0x72, 0x70, 0x63, 0x2e, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xbd, 0x01, 0x0a,
	0x0c, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a,
	0x0a, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0a, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x4b, 0x0a,
	0x0d, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x54, 0x79, 0x70, 0x65, 0x44, 0x61, 0x74, 0x61, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x4d, 0x6f, 0x63, 0x6b,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x54, 0x79,
	0x70, 0x65, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0d, 0x4d, 0x65, 0x64,
	0x69, 0x61, 0x54, 0x79, 0x70, 0x65, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x40, 0x0a, 0x12, 0x4d, 0x65,
	0x64, 0x69, 0x61, 0x54, 0x79, 0x70, 0x65, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x32, 0x46, 0x0a, 0x0d,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x12, 0x35, 0x0a,
	0x0b, 0x47, 0x65, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x12, 0x12, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73,
	0x1a, 0x0e, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74,
	0x22, 0x00, 0x30, 0x01, 0x42, 0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x73, 0x68, 0x6f, 0x70, 0x2f, 0x6b, 0x75, 0x73, 0x6b,
	0x2d, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x68, 0x65, 0x6c, 0x70, 0x65, 0x72, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d,
	0x65, 0x6e, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_management_proto_rawDescOnce sync.Once
	file_management_proto_rawDescData = file_management_proto_rawDesc
)

func file_management_proto_rawDescGZIP() []byte {
	file_management_proto_rawDescOnce.Do(func() {
		file_management_proto_rawDescData = protoimpl.X.CompressGZIP(file_management_proto_rawDescData)
	})
	return file_management_proto_rawDescData
}

var file_management_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_management_proto_goTypes = []interface{}{
	(*ClientParams)(nil), // 0: grpc.ClientParams
	(*Snapshot)(nil),     // 1: grpc.Snapshot
	(*MockConfig)(nil),   // 2: grpc.MockConfig
	(*MockResponse)(nil), // 3: grpc.MockResponse
	nil,                  // 4: grpc.MockConfig.MockResponsesEntry
	nil,                  // 5: grpc.MockResponse.MediaTypeDataEntry
}
var file_management_proto_depIdxs = []int32{
	2, // 0: grpc.Snapshot.mockConfig:type_name -> grpc.MockConfig
	4, // 1: grpc.MockConfig.MockResponses:type_name -> grpc.MockConfig.MockResponsesEntry
	5, // 2: grpc.MockResponse.MediaTypeData:type_name -> grpc.MockResponse.MediaTypeDataEntry
	3, // 3: grpc.MockConfig.MockResponsesEntry.value:type_name -> grpc.MockResponse
	0, // 4: grpc.ConfigManager.GetSnapshot:input_type -> grpc.ClientParams
	1, // 5: grpc.ConfigManager.GetSnapshot:output_type -> grpc.Snapshot
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_management_proto_init() }
func file_management_proto_init() {
	if File_management_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_management_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientParams); i {
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
		file_management_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Snapshot); i {
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
		file_management_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MockConfig); i {
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
		file_management_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MockResponse); i {
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
			RawDescriptor: file_management_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_management_proto_goTypes,
		DependencyIndexes: file_management_proto_depIdxs,
		MessageInfos:      file_management_proto_msgTypes,
	}.Build()
	File_management_proto = out.File
	file_management_proto_rawDesc = nil
	file_management_proto_goTypes = nil
	file_management_proto_depIdxs = nil
}
