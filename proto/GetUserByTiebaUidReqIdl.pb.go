// tbclient.GetUserByTiebaUid.GetUserByTiebaUidReqIdl

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.12.4
// source: GetUserByTiebaUidReqIdl.proto

package proto

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

type CommonReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	XClientType           int32   `protobuf:"varint,1,opt,name=_client_type,json=ClientType,proto3" json:"_client_type,omitempty"`
	XClientVersion        string  `protobuf:"bytes,2,opt,name=_client_version,json=ClientVersion,proto3" json:"_client_version,omitempty"`
	XClientId             string  `protobuf:"bytes,3,opt,name=_client_id,json=ClientId,proto3" json:"_client_id,omitempty"`
	XPhoneImei            string  `protobuf:"bytes,5,opt,name=_phone_imei,json=PhoneImei,proto3" json:"_phone_imei,omitempty"`
	XFrom                 string  `protobuf:"bytes,6,opt,name=_from,json=From,proto3" json:"_from,omitempty"`
	Cuid                  string  `protobuf:"bytes,7,opt,name=cuid,proto3" json:"cuid,omitempty"`
	XTimestamp            int64   `protobuf:"varint,8,opt,name=_timestamp,json=Timestamp,proto3" json:"_timestamp,omitempty"`
	Model                 string  `protobuf:"bytes,9,opt,name=model,proto3" json:"model,omitempty"`
	BDUSS                 string  `protobuf:"bytes,10,opt,name=BDUSS,proto3" json:"BDUSS,omitempty"`
	Tbs                   string  `protobuf:"bytes,11,opt,name=tbs,proto3" json:"tbs,omitempty"`
	NetType               int32   `protobuf:"varint,12,opt,name=net_type,json=netType,proto3" json:"net_type,omitempty"`
	Pversion              string  `protobuf:"bytes,24,opt,name=pversion,proto3" json:"pversion,omitempty"`
	XOsVersion            string  `protobuf:"bytes,25,opt,name=_os_version,json=OsVersion,proto3" json:"_os_version,omitempty"`
	Brand                 string  `protobuf:"bytes,26,opt,name=brand,proto3" json:"brand,omitempty"`
	LegoLibVersion        string  `protobuf:"bytes,28,opt,name=lego_lib_version,json=legoLibVersion,proto3" json:"lego_lib_version,omitempty"`
	Applist               string  `protobuf:"bytes,29,opt,name=applist,proto3" json:"applist,omitempty"`
	Stoken                string  `protobuf:"bytes,30,opt,name=stoken,proto3" json:"stoken,omitempty"`
	ZId                   string  `protobuf:"bytes,31,opt,name=z_id,json=zId,proto3" json:"z_id,omitempty"`
	CuidGalaxy2           string  `protobuf:"bytes,32,opt,name=cuid_galaxy2,json=cuidGalaxy2,proto3" json:"cuid_galaxy2,omitempty"`
	CuidGid               string  `protobuf:"bytes,33,opt,name=cuid_gid,json=cuidGid,proto3" json:"cuid_gid,omitempty"`
	C3Aid                 string  `protobuf:"bytes,35,opt,name=c3_aid,json=c3Aid,proto3" json:"c3_aid,omitempty"`
	SampleId              string  `protobuf:"bytes,36,opt,name=sample_id,json=sampleId,proto3" json:"sample_id,omitempty"`
	ScrW                  int32   `protobuf:"varint,37,opt,name=scr_w,json=scrW,proto3" json:"scr_w,omitempty"`
	ScrH                  int32   `protobuf:"varint,38,opt,name=scr_h,json=scrH,proto3" json:"scr_h,omitempty"`
	ScrDip                float64 `protobuf:"fixed64,39,opt,name=scr_dip,json=scrDip,proto3" json:"scr_dip,omitempty"`
	QType                 int32   `protobuf:"varint,40,opt,name=q_type,json=qType,proto3" json:"q_type,omitempty"`
	IsTeenager            int32   `protobuf:"varint,41,opt,name=is_teenager,json=isTeenager,proto3" json:"is_teenager,omitempty"`
	SdkVer                string  `protobuf:"bytes,42,opt,name=sdk_ver,json=sdkVer,proto3" json:"sdk_ver,omitempty"`
	FrameworkVer          string  `protobuf:"bytes,43,opt,name=framework_ver,json=frameworkVer,proto3" json:"framework_ver,omitempty"`
	NawsGameVer           string  `protobuf:"bytes,44,opt,name=naws_game_ver,json=nawsGameVer,proto3" json:"naws_game_ver,omitempty"`
	ActiveTimestamp       int64   `protobuf:"varint,49,opt,name=active_timestamp,json=activeTimestamp,proto3" json:"active_timestamp,omitempty"`
	FirstInstallTime      int64   `protobuf:"varint,50,opt,name=first_install_time,json=firstInstallTime,proto3" json:"first_install_time,omitempty"`
	LastUpdateTime        int64   `protobuf:"varint,51,opt,name=last_update_time,json=lastUpdateTime,proto3" json:"last_update_time,omitempty"`
	EventDay              string  `protobuf:"bytes,53,opt,name=event_day,json=eventDay,proto3" json:"event_day,omitempty"`
	AndroidId             string  `protobuf:"bytes,54,opt,name=android_id,json=androidId,proto3" json:"android_id,omitempty"`
	Cmode                 int32   `protobuf:"varint,55,opt,name=cmode,proto3" json:"cmode,omitempty"`
	StartScheme           string  `protobuf:"bytes,56,opt,name=start_scheme,json=startScheme,proto3" json:"start_scheme,omitempty"`
	StartType             int32   `protobuf:"varint,57,opt,name=start_type,json=startType,proto3" json:"start_type,omitempty"`
	Idfv                  string  `protobuf:"bytes,60,opt,name=idfv,proto3" json:"idfv,omitempty"`
	Extra                 string  `protobuf:"bytes,61,opt,name=extra,proto3" json:"extra,omitempty"`
	UserAgent             string  `protobuf:"bytes,62,opt,name=user_agent,json=userAgent,proto3" json:"user_agent,omitempty"`
	PersonalizedRecSwitch int32   `protobuf:"varint,63,opt,name=personalized_rec_switch,json=personalizedRecSwitch,proto3" json:"personalized_rec_switch,omitempty"`
	DeviceScore           string  `protobuf:"bytes,70,opt,name=device_score,json=deviceScore,proto3" json:"device_score,omitempty"`
}

func (x *CommonReq) Reset() {
	*x = CommonReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GetUserByTiebaUidReqIdl_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommonReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommonReq) ProtoMessage() {}

func (x *CommonReq) ProtoReflect() protoreflect.Message {
	mi := &file_GetUserByTiebaUidReqIdl_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommonReq.ProtoReflect.Descriptor instead.
func (*CommonReq) Descriptor() ([]byte, []int) {
	return file_GetUserByTiebaUidReqIdl_proto_rawDescGZIP(), []int{0}
}

func (x *CommonReq) GetXClientType() int32 {
	if x != nil {
		return x.XClientType
	}
	return 0
}

func (x *CommonReq) GetXClientVersion() string {
	if x != nil {
		return x.XClientVersion
	}
	return ""
}

func (x *CommonReq) GetXClientId() string {
	if x != nil {
		return x.XClientId
	}
	return ""
}

func (x *CommonReq) GetXPhoneImei() string {
	if x != nil {
		return x.XPhoneImei
	}
	return ""
}

func (x *CommonReq) GetXFrom() string {
	if x != nil {
		return x.XFrom
	}
	return ""
}

func (x *CommonReq) GetCuid() string {
	if x != nil {
		return x.Cuid
	}
	return ""
}

func (x *CommonReq) GetXTimestamp() int64 {
	if x != nil {
		return x.XTimestamp
	}
	return 0
}

func (x *CommonReq) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *CommonReq) GetBDUSS() string {
	if x != nil {
		return x.BDUSS
	}
	return ""
}

func (x *CommonReq) GetTbs() string {
	if x != nil {
		return x.Tbs
	}
	return ""
}

func (x *CommonReq) GetNetType() int32 {
	if x != nil {
		return x.NetType
	}
	return 0
}

func (x *CommonReq) GetPversion() string {
	if x != nil {
		return x.Pversion
	}
	return ""
}

func (x *CommonReq) GetXOsVersion() string {
	if x != nil {
		return x.XOsVersion
	}
	return ""
}

func (x *CommonReq) GetBrand() string {
	if x != nil {
		return x.Brand
	}
	return ""
}

func (x *CommonReq) GetLegoLibVersion() string {
	if x != nil {
		return x.LegoLibVersion
	}
	return ""
}

func (x *CommonReq) GetApplist() string {
	if x != nil {
		return x.Applist
	}
	return ""
}

func (x *CommonReq) GetStoken() string {
	if x != nil {
		return x.Stoken
	}
	return ""
}

func (x *CommonReq) GetZId() string {
	if x != nil {
		return x.ZId
	}
	return ""
}

func (x *CommonReq) GetCuidGalaxy2() string {
	if x != nil {
		return x.CuidGalaxy2
	}
	return ""
}

func (x *CommonReq) GetCuidGid() string {
	if x != nil {
		return x.CuidGid
	}
	return ""
}

func (x *CommonReq) GetC3Aid() string {
	if x != nil {
		return x.C3Aid
	}
	return ""
}

func (x *CommonReq) GetSampleId() string {
	if x != nil {
		return x.SampleId
	}
	return ""
}

func (x *CommonReq) GetScrW() int32 {
	if x != nil {
		return x.ScrW
	}
	return 0
}

func (x *CommonReq) GetScrH() int32 {
	if x != nil {
		return x.ScrH
	}
	return 0
}

func (x *CommonReq) GetScrDip() float64 {
	if x != nil {
		return x.ScrDip
	}
	return 0
}

func (x *CommonReq) GetQType() int32 {
	if x != nil {
		return x.QType
	}
	return 0
}

func (x *CommonReq) GetIsTeenager() int32 {
	if x != nil {
		return x.IsTeenager
	}
	return 0
}

func (x *CommonReq) GetSdkVer() string {
	if x != nil {
		return x.SdkVer
	}
	return ""
}

func (x *CommonReq) GetFrameworkVer() string {
	if x != nil {
		return x.FrameworkVer
	}
	return ""
}

func (x *CommonReq) GetNawsGameVer() string {
	if x != nil {
		return x.NawsGameVer
	}
	return ""
}

func (x *CommonReq) GetActiveTimestamp() int64 {
	if x != nil {
		return x.ActiveTimestamp
	}
	return 0
}

func (x *CommonReq) GetFirstInstallTime() int64 {
	if x != nil {
		return x.FirstInstallTime
	}
	return 0
}

func (x *CommonReq) GetLastUpdateTime() int64 {
	if x != nil {
		return x.LastUpdateTime
	}
	return 0
}

func (x *CommonReq) GetEventDay() string {
	if x != nil {
		return x.EventDay
	}
	return ""
}

func (x *CommonReq) GetAndroidId() string {
	if x != nil {
		return x.AndroidId
	}
	return ""
}

func (x *CommonReq) GetCmode() int32 {
	if x != nil {
		return x.Cmode
	}
	return 0
}

func (x *CommonReq) GetStartScheme() string {
	if x != nil {
		return x.StartScheme
	}
	return ""
}

func (x *CommonReq) GetStartType() int32 {
	if x != nil {
		return x.StartType
	}
	return 0
}

func (x *CommonReq) GetIdfv() string {
	if x != nil {
		return x.Idfv
	}
	return ""
}

func (x *CommonReq) GetExtra() string {
	if x != nil {
		return x.Extra
	}
	return ""
}

func (x *CommonReq) GetUserAgent() string {
	if x != nil {
		return x.UserAgent
	}
	return ""
}

func (x *CommonReq) GetPersonalizedRecSwitch() int32 {
	if x != nil {
		return x.PersonalizedRecSwitch
	}
	return 0
}

func (x *CommonReq) GetDeviceScore() string {
	if x != nil {
		return x.DeviceScore
	}
	return ""
}

type GetUserByTiebaUidReqIdl struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data *GetUserByTiebaUidReqIdl_DataReq `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *GetUserByTiebaUidReqIdl) Reset() {
	*x = GetUserByTiebaUidReqIdl{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GetUserByTiebaUidReqIdl_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserByTiebaUidReqIdl) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserByTiebaUidReqIdl) ProtoMessage() {}

func (x *GetUserByTiebaUidReqIdl) ProtoReflect() protoreflect.Message {
	mi := &file_GetUserByTiebaUidReqIdl_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserByTiebaUidReqIdl.ProtoReflect.Descriptor instead.
func (*GetUserByTiebaUidReqIdl) Descriptor() ([]byte, []int) {
	return file_GetUserByTiebaUidReqIdl_proto_rawDescGZIP(), []int{1}
}

func (x *GetUserByTiebaUidReqIdl) GetData() *GetUserByTiebaUidReqIdl_DataReq {
	if x != nil {
		return x.Data
	}
	return nil
}

type GetUserByTiebaUidReqIdl_DataReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Common   *CommonReq `protobuf:"bytes,1,opt,name=common,proto3" json:"common,omitempty"`
	TiebaUid string     `protobuf:"bytes,2,opt,name=tieba_uid,json=tiebaUid,proto3" json:"tieba_uid,omitempty"`
}

func (x *GetUserByTiebaUidReqIdl_DataReq) Reset() {
	*x = GetUserByTiebaUidReqIdl_DataReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GetUserByTiebaUidReqIdl_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserByTiebaUidReqIdl_DataReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserByTiebaUidReqIdl_DataReq) ProtoMessage() {}

func (x *GetUserByTiebaUidReqIdl_DataReq) ProtoReflect() protoreflect.Message {
	mi := &file_GetUserByTiebaUidReqIdl_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserByTiebaUidReqIdl_DataReq.ProtoReflect.Descriptor instead.
func (*GetUserByTiebaUidReqIdl_DataReq) Descriptor() ([]byte, []int) {
	return file_GetUserByTiebaUidReqIdl_proto_rawDescGZIP(), []int{1, 0}
}

func (x *GetUserByTiebaUidReqIdl_DataReq) GetCommon() *CommonReq {
	if x != nil {
		return x.Common
	}
	return nil
}

func (x *GetUserByTiebaUidReqIdl_DataReq) GetTiebaUid() string {
	if x != nil {
		return x.TiebaUid
	}
	return ""
}

var File_GetUserByTiebaUidReqIdl_proto protoreflect.FileDescriptor

var file_GetUserByTiebaUidReqIdl_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x42, 0x79, 0x54, 0x69, 0x65, 0x62, 0x61,
	0x55, 0x69, 0x64, 0x52, 0x65, 0x71, 0x49, 0x64, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xff, 0x09, 0x0a, 0x09, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x20, 0x0a,
	0x0c, 0x5f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0a, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x26, 0x0a, 0x0f, 0x5f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x0a, 0x5f, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x43, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1e, 0x0a, 0x0b, 0x5f, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x5f,
	0x69, 0x6d, 0x65, 0x69, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x50, 0x68, 0x6f, 0x6e,
	0x65, 0x49, 0x6d, 0x65, 0x69, 0x12, 0x13, 0x0a, 0x05, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x46, 0x72, 0x6f, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x75,
	0x69, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x75, 0x69, 0x64, 0x12, 0x1d,
	0x0a, 0x0a, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x09, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x14, 0x0a,
	0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x42, 0x44, 0x55, 0x53, 0x53, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x42, 0x44, 0x55, 0x53, 0x53, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x62, 0x73,
	0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x62, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x6e,
	0x65, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x6e,
	0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x18, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x1e, 0x0a, 0x0b, 0x5f, 0x6f, 0x73, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x19, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x4f, 0x73, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x72, 0x61, 0x6e, 0x64, 0x18, 0x1a, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x62, 0x72, 0x61, 0x6e, 0x64, 0x12, 0x28, 0x0a, 0x10, 0x6c, 0x65, 0x67, 0x6f,
	0x5f, 0x6c, 0x69, 0x62, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x1c, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0e, 0x6c, 0x65, 0x67, 0x6f, 0x4c, 0x69, 0x62, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x1d, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x1e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74,
	0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x11, 0x0a, 0x04, 0x7a, 0x5f, 0x69, 0x64, 0x18, 0x1f, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x7a, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x75, 0x69, 0x64, 0x5f,
	0x67, 0x61, 0x6c, 0x61, 0x78, 0x79, 0x32, 0x18, 0x20, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63,
	0x75, 0x69, 0x64, 0x47, 0x61, 0x6c, 0x61, 0x78, 0x79, 0x32, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x75,
	0x69, 0x64, 0x5f, 0x67, 0x69, 0x64, 0x18, 0x21, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x75,
	0x69, 0x64, 0x47, 0x69, 0x64, 0x12, 0x15, 0x0a, 0x06, 0x63, 0x33, 0x5f, 0x61, 0x69, 0x64, 0x18,
	0x23, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x63, 0x33, 0x41, 0x69, 0x64, 0x12, 0x1b, 0x0a, 0x09,
	0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x24, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x49, 0x64, 0x12, 0x13, 0x0a, 0x05, 0x73, 0x63, 0x72,
	0x5f, 0x77, 0x18, 0x25, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73, 0x63, 0x72, 0x57, 0x12, 0x13,
	0x0a, 0x05, 0x73, 0x63, 0x72, 0x5f, 0x68, 0x18, 0x26, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73,
	0x63, 0x72, 0x48, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x63, 0x72, 0x5f, 0x64, 0x69, 0x70, 0x18, 0x27,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x06, 0x73, 0x63, 0x72, 0x44, 0x69, 0x70, 0x12, 0x15, 0x0a, 0x06,
	0x71, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x28, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x71, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x73, 0x5f, 0x74, 0x65, 0x65, 0x6e, 0x61, 0x67,
	0x65, 0x72, 0x18, 0x29, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x69, 0x73, 0x54, 0x65, 0x65, 0x6e,
	0x61, 0x67, 0x65, 0x72, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x64, 0x6b, 0x5f, 0x76, 0x65, 0x72, 0x18,
	0x2a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x64, 0x6b, 0x56, 0x65, 0x72, 0x12, 0x23, 0x0a,
	0x0d, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x76, 0x65, 0x72, 0x18, 0x2b,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x77, 0x6f, 0x72, 0x6b, 0x56,
	0x65, 0x72, 0x12, 0x22, 0x0a, 0x0d, 0x6e, 0x61, 0x77, 0x73, 0x5f, 0x67, 0x61, 0x6d, 0x65, 0x5f,
	0x76, 0x65, 0x72, 0x18, 0x2c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6e, 0x61, 0x77, 0x73, 0x47,
	0x61, 0x6d, 0x65, 0x56, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x10, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x31, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0f, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x2c, 0x0a, 0x12, 0x66, 0x69, 0x72, 0x73, 0x74, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6c, 0x6c, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x32, 0x20, 0x01, 0x28, 0x03, 0x52, 0x10, 0x66,
	0x69, 0x72, 0x73, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x28, 0x0a, 0x10, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x33, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0e, 0x6c, 0x61, 0x73, 0x74, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x5f, 0x64, 0x61, 0x79, 0x18, 0x35, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x44, 0x61, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x6e, 0x64, 0x72, 0x6f, 0x69,
	0x64, 0x5f, 0x69, 0x64, 0x18, 0x36, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x6e, 0x64, 0x72,
	0x6f, 0x69, 0x64, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x37,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x63, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x65, 0x18, 0x38, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x73, 0x74, 0x61, 0x72, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x65, 0x12, 0x1d,
	0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x39, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x69, 0x64, 0x66, 0x76, 0x18, 0x3c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x69, 0x64, 0x66,
	0x76, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x18, 0x3d, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x61, 0x67, 0x65, 0x6e, 0x74, 0x18, 0x3e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x75, 0x73, 0x65,
	0x72, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x12, 0x36, 0x0a, 0x17, 0x70, 0x65, 0x72, 0x73, 0x6f, 0x6e,
	0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x63, 0x5f, 0x73, 0x77, 0x69, 0x74, 0x63,
	0x68, 0x18, 0x3f, 0x20, 0x01, 0x28, 0x05, 0x52, 0x15, 0x70, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x61,
	0x6c, 0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x63, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x12, 0x21,
	0x0a, 0x0c, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x46,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x53, 0x63, 0x6f, 0x72,
	0x65, 0x22, 0x9b, 0x01, 0x0a, 0x17, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x42, 0x79, 0x54,
	0x69, 0x65, 0x62, 0x61, 0x55, 0x69, 0x64, 0x52, 0x65, 0x71, 0x49, 0x64, 0x6c, 0x12, 0x34, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x47, 0x65,
	0x74, 0x55, 0x73, 0x65, 0x72, 0x42, 0x79, 0x54, 0x69, 0x65, 0x62, 0x61, 0x55, 0x69, 0x64, 0x52,
	0x65, 0x71, 0x49, 0x64, 0x6c, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x1a, 0x4a, 0x0a, 0x07, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x12, 0x22,
	0x0a, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a,
	0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x52, 0x06, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x69, 0x65, 0x62, 0x61, 0x5f, 0x75, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x69, 0x65, 0x62, 0x61, 0x55, 0x69, 0x64, 0x42,
	0x26, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x42, 0x41,
	0x4e, 0x4b, 0x41, 0x32, 0x30, 0x31, 0x37, 0x2f, 0x74, 0x62, 0x73, 0x69, 0x67, 0x6e, 0x5f, 0x67,
	0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_GetUserByTiebaUidReqIdl_proto_rawDescOnce sync.Once
	file_GetUserByTiebaUidReqIdl_proto_rawDescData = file_GetUserByTiebaUidReqIdl_proto_rawDesc
)

func file_GetUserByTiebaUidReqIdl_proto_rawDescGZIP() []byte {
	file_GetUserByTiebaUidReqIdl_proto_rawDescOnce.Do(func() {
		file_GetUserByTiebaUidReqIdl_proto_rawDescData = protoimpl.X.CompressGZIP(file_GetUserByTiebaUidReqIdl_proto_rawDescData)
	})
	return file_GetUserByTiebaUidReqIdl_proto_rawDescData
}

var file_GetUserByTiebaUidReqIdl_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_GetUserByTiebaUidReqIdl_proto_goTypes = []interface{}{
	(*CommonReq)(nil),                       // 0: CommonReq
	(*GetUserByTiebaUidReqIdl)(nil),         // 1: GetUserByTiebaUidReqIdl
	(*GetUserByTiebaUidReqIdl_DataReq)(nil), // 2: GetUserByTiebaUidReqIdl.DataReq
}
var file_GetUserByTiebaUidReqIdl_proto_depIdxs = []int32{
	2, // 0: GetUserByTiebaUidReqIdl.data:type_name -> GetUserByTiebaUidReqIdl.DataReq
	0, // 1: GetUserByTiebaUidReqIdl.DataReq.common:type_name -> CommonReq
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_GetUserByTiebaUidReqIdl_proto_init() }
func file_GetUserByTiebaUidReqIdl_proto_init() {
	if File_GetUserByTiebaUidReqIdl_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_GetUserByTiebaUidReqIdl_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommonReq); i {
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
		file_GetUserByTiebaUidReqIdl_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUserByTiebaUidReqIdl); i {
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
		file_GetUserByTiebaUidReqIdl_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUserByTiebaUidReqIdl_DataReq); i {
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
			RawDescriptor: file_GetUserByTiebaUidReqIdl_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_GetUserByTiebaUidReqIdl_proto_goTypes,
		DependencyIndexes: file_GetUserByTiebaUidReqIdl_proto_depIdxs,
		MessageInfos:      file_GetUserByTiebaUidReqIdl_proto_msgTypes,
	}.Build()
	File_GetUserByTiebaUidReqIdl_proto = out.File
	file_GetUserByTiebaUidReqIdl_proto_rawDesc = nil
	file_GetUserByTiebaUidReqIdl_proto_goTypes = nil
	file_GetUserByTiebaUidReqIdl_proto_depIdxs = nil
}
