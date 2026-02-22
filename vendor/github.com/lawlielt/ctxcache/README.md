# ctxcache
`ctxcache` is a lightweight, efficient context caching library designed specifically for caching data within the lifecycle of a request in Go applications, particularly useful in environments using gRPC.

## Features

* **Context-Specific Caching**: Tie cache entries to request contexts, ensuring that cache data does not bleed over across requests.
* **Automatic Cleanup**: Cache entries are automatically cleaned up when the context is cancelled or times out, helping prevent memory leaks.
* **Concurrency-Safe**: Utilize Go's concurrency features to make the cache safe to use across multiple goroutines without additional synchronization.
* **Easy Integration**: Designed to be easily integrated with gRPC interceptors or any middleware-like architecture.

## Installation

To install `ctxcache`, you need a working Go environment. You can then install `ctxcache` using `go get`:
```
go get github.com/lawlielt/ctxcache
```

## Usage


## Contributing

We welcome contributions from the community! If you would like to contribute to `ctxcache`, please fork the repository, make your changes, and submit a pull request.

## Support and Community

If you have any questions or need help integrating `ctxcache` into your application, please create an issue in the GitHub repository.

## License

ctxcache is released under the MIT License. See the LICENSE file in the repository for more details.
