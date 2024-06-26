// tbclient.GetBawuInfo.GetBawuInfoReqIdl

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.12.4
// source: GetBawuInfoReqIdl.proto

package __

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

type GetBawuInfoReqIdl struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data *GetBawuInfoReqIdl_DataReq `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *GetBawuInfoReqIdl) Reset() {
	*x = GetBawuInfoReqIdl{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GetBawuInfoReqIdl_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBawuInfoReqIdl) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBawuInfoReqIdl) ProtoMessage() {}

func (x *GetBawuInfoReqIdl) ProtoReflect() protoreflect.Message {
	mi := &file_GetBawuInfoReqIdl_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBawuInfoReqIdl.ProtoReflect.Descriptor instead.
func (*GetBawuInfoReqIdl) Descriptor() ([]byte, []int) {
	return file_GetBawuInfoReqIdl_proto_rawDescGZIP(), []int{0}
}

func (x *GetBawuInfoReqIdl) GetData() *GetBawuInfoReqIdl_DataReq {
	if x != nil {
		return x.Data
	}
	return nil
}

type GetBawuInfoReqIdl_DataReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Common *CommonReq `protobuf:"bytes,1,opt,name=common,proto3" json:"common,omitempty"`
	Fid    uint64     `protobuf:"varint,2,opt,name=fid,proto3" json:"fid,omitempty"`
}

func (x *GetBawuInfoReqIdl_DataReq) Reset() {
	*x = GetBawuInfoReqIdl_DataReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GetBawuInfoReqIdl_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBawuInfoReqIdl_DataReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBawuInfoReqIdl_DataReq) ProtoMessage() {}

func (x *GetBawuInfoReqIdl_DataReq) ProtoReflect() protoreflect.Message {
	mi := &file_GetBawuInfoReqIdl_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBawuInfoReqIdl_DataReq.ProtoReflect.Descriptor instead.
func (*GetBawuInfoReqIdl_DataReq) Descriptor() ([]byte, []int) {
	return file_GetBawuInfoReqIdl_proto_rawDescGZIP(), []int{0, 0}
}

func (x *GetBawuInfoReqIdl_DataReq) GetCommon() *CommonReq {
	if x != nil {
		return x.Common
	}
	return nil
}

func (x *GetBawuInfoReqIdl_DataReq) GetFid() uint64 {
	if x != nil {
		return x.Fid
	}
	return 0
}

var File_GetBawuInfoReqIdl_proto protoreflect.FileDescriptor

var file_GetBawuInfoReqIdl_proto_rawDesc = []byte{
	0x0a, 0x17, 0x47, 0x65, 0x74, 0x42, 0x61, 0x77, 0x75, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71,
	0x49, 0x64, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x43, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x52, 0x65, 0x71, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x84, 0x01, 0x0a, 0x11, 0x47,
	0x65, 0x74, 0x42, 0x61, 0x77, 0x75, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x49, 0x64, 0x6c,
	0x12, 0x2e, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x47, 0x65, 0x74, 0x42, 0x61, 0x77, 0x75, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x49,
	0x64, 0x6c, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x1a, 0x3f, 0x0a, 0x07, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x12, 0x22, 0x0a, 0x06, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x43, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x52, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x12,
	0x10, 0x0a, 0x03, 0x66, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x66, 0x69,
	0x64, 0x42, 0x04, 0x5a, 0x02, 0x2e, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_GetBawuInfoReqIdl_proto_rawDescOnce sync.Once
	file_GetBawuInfoReqIdl_proto_rawDescData = file_GetBawuInfoReqIdl_proto_rawDesc
)

func file_GetBawuInfoReqIdl_proto_rawDescGZIP() []byte {
	file_GetBawuInfoReqIdl_proto_rawDescOnce.Do(func() {
		file_GetBawuInfoReqIdl_proto_rawDescData = protoimpl.X.CompressGZIP(file_GetBawuInfoReqIdl_proto_rawDescData)
	})
	return file_GetBawuInfoReqIdl_proto_rawDescData
}

var file_GetBawuInfoReqIdl_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_GetBawuInfoReqIdl_proto_goTypes = []interface{}{
	(*GetBawuInfoReqIdl)(nil),         // 0: GetBawuInfoReqIdl
	(*GetBawuInfoReqIdl_DataReq)(nil), // 1: GetBawuInfoReqIdl.DataReq
	(*CommonReq)(nil),                 // 2: CommonReq
}
var file_GetBawuInfoReqIdl_proto_depIdxs = []int32{
	1, // 0: GetBawuInfoReqIdl.data:type_name -> GetBawuInfoReqIdl.DataReq
	2, // 1: GetBawuInfoReqIdl.DataReq.common:type_name -> CommonReq
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_GetBawuInfoReqIdl_proto_init() }
func file_GetBawuInfoReqIdl_proto_init() {
	if File_GetBawuInfoReqIdl_proto != nil {
		return
	}
	file_CommonReq_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_GetBawuInfoReqIdl_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBawuInfoReqIdl); i {
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
		file_GetBawuInfoReqIdl_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBawuInfoReqIdl_DataReq); i {
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
			RawDescriptor: file_GetBawuInfoReqIdl_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_GetBawuInfoReqIdl_proto_goTypes,
		DependencyIndexes: file_GetBawuInfoReqIdl_proto_depIdxs,
		MessageInfos:      file_GetBawuInfoReqIdl_proto_msgTypes,
	}.Build()
	File_GetBawuInfoReqIdl_proto = out.File
	file_GetBawuInfoReqIdl_proto_rawDesc = nil
	file_GetBawuInfoReqIdl_proto_goTypes = nil
	file_GetBawuInfoReqIdl_proto_depIdxs = nil
}
