
# Resource Lock Package

The `rl` package provides a flexible resource locking mechanism that allows you to choose between a local Go mutex implementation or a Redis-based implementation. 
This is useful for ensuring mutual exclusion when accessing shared resources in distributed systems or in single-node applications, adapting to requirements changes.

## Installation

First, install the package using `go get`:

```sh
go get https://github.com/dacalin/resource_lock
```

## Usage

You can use the `ResourceLockBuilder` to create a lock based on your requirements. The builder supports configuring a local mutex lock or a Redis-based lock.

### Example: Using Local Lock

```go
package main

import (
    "fmt"
    "time"
    "github.com/dacalin/resource_lock/rl"
)

func main() {
    lock := rl.New(rl.Local).
        WithCleanMemInterval(5000). // Set clean memory interval to 5000 milliseconds
        Build()

    resourceID := "exampleResource"

    lock.Lock(resourceID)
    fmt.Println("Locked resource with Local Lock")

    // Simulate some work with the resource
    time.Sleep(500 * time.Millisecond)

    lock.Unlock(resourceID)
    fmt.Println("Unlocked resource with Local Lock")
}
```

### Example: Using Redis Lock

Make sure you have a Redis server running and accessible.

```go
package main

import (
    "fmt"
    "os"
    "time"
    "github.com/dacalin/resource_lock/rl"
)

func main() {
    redisHost := os.Getenv("REDIS_HOST")
    redisPort := os.Getenv("REDIS_PORT")
    redisPoolSize := 10
    cleanMemMilis := 5000 // Set clean memory interval to 5000 milliseconds

    lock := rl.New(rl.Redis).
        WithRedisConfig(redisHost, redisPort, redisPoolSize).
        WithCleanMemInterval(cleanMemMilis).
        Build()

    resourceID := "exampleResource"

    lock.Lock(resourceID)
    fmt.Println("Locked resource with Redis Lock")

    // Simulate some work with the resource
    time.Sleep(2 * time.Second)

    lock.Unlock(resourceID)
    fmt.Println("Unlocked resource with Redis Lock")
}
```

## API Reference

### `ResourceLockBuilder`

- `New(lockType LockType) *ResourceLockBuilder`
  - Creates a new `ResourceLockBuilder` with the specified lock type (`Local` or `Redis`).

- `(*ResourceLockBuilder) WithRedisConfig(host, port string, poolSize int) *ResourceLockBuilder`
  - Configures the Redis connection parameters.
  - `host`: Redis server host.
  - `port`: Redis server port.
  - `poolSize`: Number of connections in the Redis pool.

- `(*ResourceLockBuilder) WithCleanMemInterval(milis int64) *ResourceLockBuilder`
  - Sets the interval for cleaning up expired locks.
  - `milis`: Interval in milliseconds.

- `(*ResourceLockBuilder) Build() _i_resource_lock.IResourceLock`
  - Builds and returns the configured lock instance.

### `LockType`

- `Local`
  - Use a local Go mutex for locking.

- `Redis`
  - Use Redis for distributed locking.

## Environment Variables

When using the Redis implementation, you can set the following environment variables:

- `REDIS_HOST`: The host of the Redis server.
- `REDIS_PORT`: The port of the Redis server.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License.
