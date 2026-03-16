package redis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pthethanh/nano/cache"
	cacheRedis "github.com/pthethanh/nano/plugins/cache/redis"
)

func TestCache(t *testing.T) {
	s := miniredis.RunT(t)

	var c cache.Cacher[string, []byte] = cacheRedis.New(
		cacheRedis.Address[string, []byte](s.Addr()),
		cacheRedis.CodecOption[string, []byte](cacheRedis.BytesCodec{}),
	)
	if err := c.Open(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer c.Close(context.Background())

	if err := c.Set(context.Background(), "k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if v, err := c.Get(context.Background(), "k"); err != nil || string(v) != "v" {
		t.Fatalf("got result=%v, err=%v, want result=%v, err=%v", string(v), err, "v", nil)
	}
	if _, err := c.Get(context.Background(), "missing"); err != cache.ErrNotFound {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrNotFound)
	}
}

func TestCacheTimeout(t *testing.T) {
	s := miniredis.RunT(t)

	c := cacheRedis.New[string, []byte](
		cacheRedis.Address[string, []byte](s.Addr()),
		cacheRedis.CodecOption[string, []byte](cacheRedis.BytesCodec{}),
	)
	if err := c.Open(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer c.Close(context.Background())

	if err := c.Set(context.Background(), "k", []byte("v"), cache.TTL(time.Second)); err != nil {
		t.Fatal(err)
	}
	s.FastForward(1500 * time.Millisecond)
	if _, err := c.Get(context.Background(), "k"); err != cache.ErrNotFound {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrNotFound)
	}
}

func TestCacheDelete(t *testing.T) {
	s := miniredis.RunT(t)

	c := cacheRedis.New[string, []byte](
		cacheRedis.Address[string, []byte](s.Addr()),
		cacheRedis.CodecOption[string, []byte](cacheRedis.BytesCodec{}),
	)
	if err := c.Open(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer c.Close(context.Background())

	if err := c.Set(context.Background(), "k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(context.Background(), "k"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(context.Background(), "k"); err != cache.ErrNotFound {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrNotFound)
	}
}

func TestInvalidState(t *testing.T) {
	c := cacheRedis.New(cacheRedis.CodecOption[string, []byte](cacheRedis.BytesCodec{}))

	if _, err := c.Get(context.Background(), "k"); err != cache.ErrInValidConnState {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrInValidConnState)
	}
	if err := c.Set(context.Background(), "k", []byte("v")); err != cache.ErrInValidConnState {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrInValidConnState)
	}
	if err := c.Delete(context.Background(), "k"); err != cache.ErrInValidConnState {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrInValidConnState)
	}
}

func TestGenericCacheWithJSONCodec(t *testing.T) {
	s := miniredis.RunT(t)

	type value struct {
		Name string
		Age  int
	}

	c := cacheRedis.New[int, value](
		cacheRedis.Address[int, value](s.Addr()),
		cacheRedis.KeyEncoder[int, value](func(k int) string { return fmt.Sprintf("person:%d", k) }),
	)
	if err := c.Open(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer c.Close(context.Background())

	want := value{Name: "alice", Age: 30}
	if err := c.Set(context.Background(), 7, want); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got result=%v, want result=%v", got, want)
	}
}
