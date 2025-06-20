package _function

import (
	"strconv"
)

type VerifyCodeStruct struct {
	VerifyCode string `json:"verify_code"`
	Value      string `json:"value"`
	ResetTime  int64  `json:"time"`
	TryTime    int64  `json:"try_time"`
	Type       string `json:"type"`
}

type VerifyCodeListType struct {
	List KV[string, *VerifyCodeStruct]
}

var VerifyCodeList VerifyCodeListType //= make(map[int32]*ResetPwdStruct)

func (list *VerifyCodeListType) StoreCode(_type string, uid int32, data *VerifyCodeStruct) {
	list.List.Store(_type+":"+strconv.Itoa(int(uid)), data, ResetPwdExpire)
}

func (list *VerifyCodeListType) LoadCode(_type string, uid int32) (*VerifyCodeStruct, bool) {
	return list.List.Load(_type + ":" + strconv.Itoa(int(uid)))
}

func (list *VerifyCodeListType) DeleteCode(_type string, uid int32) {
	list.List.Delete(_type + ":" + strconv.Itoa(int(uid)))
}
