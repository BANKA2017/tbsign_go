package _function

import (
	"strconv"
)

func init() {
	VerifyCodeList.List = NewKV[string, *VerifyCodeStruct]()
}

type VerifyCodeStruct struct {
	VerifyCode string `json:"verify_code"`
	Value      string `json:"value"`
	ResetTime  int64  `json:"time"`
	TryTime    int64  `json:"try_time"`
	Type       string `json:"type"`
	Expire     int64  `json:"expire"`
}

type VerifyCodeListType struct {
	List *KV[string, *VerifyCodeStruct]
}

var VerifyCodeList VerifyCodeListType //= make(map[int32]*ResetPwdStruct)

func (list *VerifyCodeListType) StoreCode(_type string, uid int32, data *VerifyCodeStruct) {
	list.List.Store(_type+":"+strconv.Itoa(int(uid)), data, ResetPwdExpire)
}

func (list *VerifyCodeListType) LoadCode(_type string, uid int32) (*VerifyCodeStruct, bool) {
	key := _type + ":" + strconv.Itoa(int(uid))
	v, s := list.List.Load(key)
	if s {
		ttl, s := list.List.TTL(key)
		if ttl > 0 && v.Expire != int64(ttl) {
			v.Expire = int64(ttl)
		}
		return v, s
	}
	return nil, false
}

func (list *VerifyCodeListType) DeleteCode(_type string, uid int32) {
	list.List.Delete(_type + ":" + strconv.Itoa(int(uid)))
}
