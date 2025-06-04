package _function

import (
	"sync"
	"time"
)

type KV[K, T any] struct {
	KV sync.Map
}

type KVStruct[T any] struct {
	Value    T     `json:"value"`
	ExpireAt int64 `json:"expire_at"`
}

func (list *KV[K, T]) Store(key K, value T, ttlSeconds int64) {
	list.KV.Store(key, &KVStruct[T]{
		Value:    value,
		ExpireAt: When(ttlSeconds == -1, -1, time.Now().Add(time.Second*time.Duration(ttlSeconds)).Unix()),
	})
}

func (list *KV[K, T]) Load(key K) (T, bool) {
	v, ok := list.KV.Load(key)

	if !ok {
		var nullValue T
		return nullValue, false
	}

	vStructed := v.(*KVStruct[T])

	if vStructed.ExpireAt > -1 && vStructed.ExpireAt < time.Now().Unix() {
		list.Delete(key)
		var nullValue T
		return nullValue, false
	}

	return vStructed.Value, true
}

func (list *KV[K, T]) LoadAndDelete(key K) (T, bool) {
	v, ok := list.KV.LoadAndDelete(key)

	if !ok {
		var nullValue T
		return nullValue, false
	}

	vStructed := v.(*KVStruct[T])

	if vStructed.ExpireAt > -1 && vStructed.ExpireAt < time.Now().Unix() {
		var nullValue T
		return nullValue, false
	}

	return vStructed.Value, true
}

func (list *KV[K, T]) Delete(key K) {
	list.KV.Delete(key)
}

func (list *KV[K, T]) Range(f func(key any, value any) bool) {
	list.KV.Range(f)
}

func (list *KV[K, T]) RemoveExpired() {
	now := time.Now().Unix()
	list.Range(func(key, value any) bool {
		expireAt := value.(*KVStruct[T]).ExpireAt
		if expireAt > -1 && expireAt < now {
			list.Delete(key.(K))
		}
		return true
	})
}
