# Build

This project requires setting up MinGW-w64.<br>

* `markdown-live-preview` depends on `webview_go`.
* `webview_go` requires CGO.
* CGO is [not compatible with MSVC](https://go.dev/wiki/cgo#windows). Setting up GCC is required.

I recommend using [WinLibs](https://winlibs.com/)

* GCC 15.1.0 with MCF threads, without LLVM/Clang/LLD/LLDB

## Resource

If you make `.syso` file, `go build` will apply it automatically.

To make `.syso` file, install [`go-winres`](https://github.com/tc-hib/go-winres).

```
go install github.com/tc-hib/go-winres@latest
```

Then run 