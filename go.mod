module github.com/traefik/plugindemowasm-http-call

go 1.22.3

toolchain go1.22.4

require (
	github.com/http-wasm/http-wasm-guest-tinygo v0.4.0
	github.com/juliens/wasm-goexport v0.0.6
	github.com/stealthrocket/net v0.2.1
)

require github.com/tetratelabs/wazero v1.7.3 // indirect

replace github.com/http-wasm/http-wasm-guest-tinygo => github.com/juliens/http-wasm-guest-tinygo v0.0.0-20240602204949-9cdd64d990eb
