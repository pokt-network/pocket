// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: codec_test.proto

package codec

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

// This structure is needed to test the codec itself
// If we did not have this structure, we'd need to `import` specific
// proto structures in order to test the `Marhsal` and `Unmarshal` functions
// See https://github.com/pokt-network/pocket/issues/231 for more details
type TestProtoStructure struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Field1 int32  `protobuf:"varint,1,opt,name=field1,proto3" json:"field1,omitempty"`
	Field2 string `protobuf:"bytes,2,opt,name=field2,proto3" json:"field2,omitempty"`
	Field3 bool   `protobuf:"varint,3,opt,name=field3,proto3" json:"field3,omitempty"`
}

func (x *TestProtoStructure) Reset() {
	*x = TestProtoStructure{}
	if protoimpl.UnsafeEnabled {
		mi := &file_codec_test_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestProtoStructure) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestProtoStructure) ProtoMessage() {}

func (x *TestProtoStructure) ProtoReflect() protoreflect.Message {
	mi := &file_codec_test_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestProtoStructure.ProtoReflect.Descriptor instead.
func (*TestProtoStructure) Descriptor() ([]byte, []int) {
	return file_codec_test_proto_rawDescGZIP(), []int{0}
}

func (x *TestProtoStructure) GetField1() int32 {
	if x != nil {
		return x.Field1
	}
	return 0
}

func (x *TestProtoStructure) GetField2() string {
	if x != nil {
		return x.Field2
	}
	return ""
}

func (x *TestProtoStructure) GetField3() bool {
	if x != nil {
		return x.Field3
	}
	return false
}

var File_codec_test_proto protoreflect.FileDescriptor

var file_codec_test_proto_rawDesc = []byte{
	0x0a, 0x10, 0x63, 0x6f, 0x64, 0x65, 0x63, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x06, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x22, 0x5c, 0x0a, 0x12, 0x54, 0x65,
	0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x75, 0x72, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x31, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x31, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c,
	0x64, 0x32, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x32,
	0x12, 0x16, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x33, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x33, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x6f, 0x6b, 0x74, 0x2d, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x2f, 0x70, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2f, 0x73, 0x68, 0x61, 0x72, 0x65,
	0x64, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_codec_test_proto_rawDescOnce sync.Once
	file_codec_test_proto_rawDescData = file_codec_test_proto_rawDesc
)

func file_codec_test_proto_rawDescGZIP() []byte {
	file_codec_test_proto_rawDescOnce.Do(func() {
		file_codec_test_proto_rawDescData = protoimpl.X.CompressGZIP(file_codec_test_proto_rawDescData)
	})
	return file_codec_test_proto_rawDescData
}

var file_codec_test_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_codec_test_proto_goTypes = []interface{}{
	(*TestProtoStructure)(nil), // 0: shared.TestProtoStructure
}
var file_codec_test_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_codec_test_proto_init() }
func file_codec_test_proto_init() {
	if File_codec_test_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_codec_test_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestProtoStructure); i {
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
			RawDescriptor: file_codec_test_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_codec_test_proto_goTypes,
		DependencyIndexes: file_codec_test_proto_depIdxs,
		MessageInfos:      file_codec_test_proto_msgTypes,
	}.Build()
	File_codec_test_proto = out.File
	file_codec_test_proto_rawDesc = nil
	file_codec_test_proto_goTypes = nil
	file_codec_test_proto_depIdxs = nil
}
