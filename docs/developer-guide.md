# Developer Guide

There's a `Makefile` in the root folder.

- Build docker images for specific architecture, supported architectures are `amd64`,`arm`,`arm64`:

```bash
ARCH=amd64 make build
```

- Build all docker images for all supported architectures.

```bash
make build
```
