package _function

import (
	"sync"
	"time"
)

type KV[K, T any] struct {
	KV sync.Map
	// SF singleflight.Group
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

// Unix timestamp
func (list *KV[K, T]) TTL(key K) (int, bool) {
	v, ok := list.KV.Load(key)

	if !ok {
		return 0, false
	}

	vStructed := v.(*KVStruct[T])

	return int(vStructed.ExpireAt), true
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

func (list *KV[K, T]) DeleteAll() {
	list.KV.Range(func(key, value any) bool {
		list.KV.Delete(key)
		return true
	})
}

func (list *KV[K, T]) Length() int {
	length := 0
	list.KV.Range(func(key, value any) bool {
		length++
		return true
	})
	return length
}

func (list *KV[K, T]) Range(f func(key K, value T) bool) {
	list.KV.Range(func(k, v any) bool {
		typedKey, ok1 := k.(K)
		typedVal, ok2 := v.(*KVStruct[T])
		if !ok1 || !ok2 {
			return true
		}
		return f(typedKey, typedVal.Value)
	})
}

func (list *KV[K, T]) RemoveExpired() {
	now := time.Now().Unix()
	list.KV.Range(func(key, value any) bool {
		expireAt := value.(*KVStruct[T]).ExpireAt
		if expireAt > -1 && expireAt < now {
			list.Delete(key.(K))
		}
		return true
	})
}
