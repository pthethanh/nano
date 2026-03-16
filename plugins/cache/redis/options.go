package redis

import goredis "github.com/redis/go-redis/v9"

type Option[K comparable, V any] func(*Cacher[K, V])

// Address configures Redis server addresses.
func Address[K comparable, V any](addrs ...string) Option[K, V] {
	return func(c *Cacher[K, V]) {
		c.opts.Addrs = append([]string(nil), addrs...)
	}
}

// Username configures Redis ACL username.
func Username[K comparable, V any](username string) Option[K, V] {
	return func(c *Cacher[K, V]) {
		c.opts.Username = username
	}
}

// Password configures Redis password.
func Password[K comparable, V any](password string) Option[K, V] {
	return func(c *Cacher[K, V]) {
		c.opts.Password = password
	}
}

// DB configures the Redis logical database.
func DB[K comparable, V any](db int) Option[K, V] {
	return func(c *Cacher[K, V]) {
		c.opts.DB = db
	}
}

// MasterName configures Redis Sentinel master name.
func MasterName[K comparable, V any](masterName string) Option[K, V] {
	return func(c *Cacher[K, V]) {
		c.opts.MasterName = masterName
	}
}

// ClientName configures the Redis client name.
func ClientName[K comparable, V any](clientName string) Option[K, V] {
	return func(c *Cacher[K, V]) {
		c.opts.ClientName = clientName
	}
}

// Options replaces the Redis universal options.
func Options[K comparable, V any](opts *goredis.UniversalOptions) Option[K, V] {
	return func(c *Cacher[K, V]) {
		if opts == nil {
			return
		}
		clone := *opts
		if opts.Addrs != nil {
			clone.Addrs = append([]string(nil), opts.Addrs...)
		}
		c.opts = &clone
	}
}

// Client injects an existing Redis client.
func Client[K comparable, V any](client goredis.UniversalClient) Option[K, V] {
	return func(c *Cacher[K, V]) {
		c.client = client
		c.managed = false
	}
}

// KeyEncoder configures how cache keys are converted into Redis keys.
func KeyEncoder[K comparable, V any](fn KeyFunc[K]) Option[K, V] {
	return func(c *Cacher[K, V]) {
		if fn != nil {
			c.keyFunc = fn
		}
	}
}

// CodecOption configures how cache values are serialized into Redis.
func CodecOption[K comparable, V any](codec Codec[V]) Option[K, V] {
	return func(c *Cacher[K, V]) {
		if codec != nil {
			c.codec = codec
		}
	}
}
