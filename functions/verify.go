package _function

import (
	"strconv"
	"sync"
)

type VerifyCodeStruct struct {
	VerifyCode string `json:"verify_code"`
	Expire     int64  `json:"expire"`
	Value      string `json:"value"`
	ResetTime  int64  `json:"time"`
	TryTime    int64  `json:"try_time"`
	Type       string `json:"type"`
}

type VerifyCodeListType struct {
	List sync.Map
}

var VerifyCodeList VerifyCodeListType //= make(map[int32]*ResetPwdStruct)

func (list *VerifyCodeListType) Store(key, value any) {
	list.List.Store(key, value)
}

func (list *VerifyCodeListType) Load(key any) (any, bool) {
	return list.List.Load(key)
}

func (list *VerifyCodeListType) Delete(key any) {
	list.List.Delete(key)
}

func (list *VerifyCodeListType) Range(f func(key any, value any) bool) {
	list.List.Range(f)
}

func (list *VerifyCodeListType) RemoveExpired() {
	list.Range(func(key, value any) bool {
		if value.(*VerifyCodeStruct).Expire < Now.Unix() {
			list.Delete(key)
		}
		return true
	})
}

func (list *VerifyCodeListType) StoreCode(_type string, uid int32, data *VerifyCodeStruct) {
	list.Store(AppendStrings(_type, ":", strconv.Itoa(int(uid))), data)
}

func (list *VerifyCodeListType) LoadCode(_type string, uid int32) (*VerifyCodeStruct, bool) {
	_code, ok := list.Load(AppendStrings(_type, ":", strconv.Itoa(int(uid))))
	if !ok {
		return nil, ok
	}
	code, ok := _code.(*VerifyCodeStruct)
	return code, ok
}

func (list *VerifyCodeListType) DeleteCode(_type string, uid int32) {
	list.Delete(AppendStrings(_type, ":", strconv.Itoa(int(uid))))
}
