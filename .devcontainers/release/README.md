# Release Containers

This region is reserved for release periods and serves julec's build for different platforms.
Please use `build.sh` to run the whole process automatically.
The `build.sh` designed for root directory of project.
Please execute `build.sh` when you are in root directory of project.

## Supported Platforms

Listed below are the unique OS-ARCH operations used in the build process.

- `linux-amd64`
- `linux-arm64`

## Preparing to Build

It is recommended that you first be in the [`root`](https://github.com/julelang/jule) directory of project to execute build script.
Images use the [julec-IR](https://github.com/julelang/julec-ir) repository. Therefore, be sure to push the version you want to release to the julec-IR repository.
