package _function

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type KV[K comparable, T any] struct {
	KV *ttlcache.Cache[K, T]
}

// type KVStruct[T any] struct {
// 	Value    T     `json:"value"`
// 	ExpireAt int64 `json:"expire_at"`
// }

func NewKV[K comparable, T any](opts ...ttlcache.Option[K, T]) *KV[K, T] {
	return &KV[K, T]{
		KV: ttlcache.New(opts...),
	}
}

func (list *KV[K, T]) Store(key K, value T, ttlSeconds int64) {
	var ttl time.Duration
	if ttlSeconds <= -1 {
		ttl = ttlcache.NoTTL
	} else {
		ttl = time.Duration(ttlSeconds) * time.Second
	}

	list.KV.Set(key, value, ttl)
}

func (list *KV[K, T]) Load(key K) (T, bool) {
	v := list.KV.Get(key)

	if v == nil {
		var nullValue T
		return nullValue, false
	}

	return v.Value(), true
}

func (list *KV[K, T]) LoadWithTTL(key K) (T, int, bool) {
	v := list.KV.Get(key)

	if v == nil {
		var nullValue T
		return nullValue, 0, false
	}

	return v.Value(), int(v.ExpiresAt().Unix()), true
}

// Unix timestamp
func (list *KV[K, T]) TTL(key K) (int, bool) {
	v := list.KV.Get(key)

	if v == nil {
		return 0, false
	}

	return int(v.TTL()), true
}

func (list *KV[K, T]) LoadAndDelete(key K) (T, bool) {
	v, _ := list.KV.GetAndDelete(key)

	if v == nil {
		var nullValue T
		return nullValue, false
	}

	return v.Value(), true
}

func (list *KV[K, T]) Delete(key K) {
	list.KV.Delete(key)
}

func (list *KV[K, T]) DeleteAll() {
	list.KV.DeleteAll()
}

func (list *KV[K, T]) Length() int {
	return list.KV.Len()
}

func (list *KV[K, T]) Range(f func(key K, value T) bool) {
	list.KV.Range(func(item *ttlcache.Item[K, T]) bool {
		if item != nil {
			return f(item.Key(), item.Value())
		}

		return true
	})
}

func (list *KV[K, T]) RemoveExpired() {
	list.KV.DeleteExpired()
}
