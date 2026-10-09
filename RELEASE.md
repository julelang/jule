# Release Guide

> [!WARNING]
> This documentation is intended for maintainers.

This documentation describes how the processes should be managed before a Jule release is published.\
A maintainer is responsible for following these processes when publishing a release.

## Preparation

- The `Version` constant in the `std/jule` package must be set correctly.
- [julec-ir](https://github.com/julelang/julec-ir) must be updated to the latest IR version, including the steps mentioned above.
- The Manual must be fully updated according to the latest release (do not forget to update the version table in the [backend](https://manual.jule.dev/compiler/backend/) section).
- The runtime malloc implementation must be updated to the latest version of the adopted memory allocator. Information regarding the adopted memory allocator is specified in the malloc implementation source file.

**Make sure the required dependencies are installed specified for the platform.**

## Build

This stage is about building the compiler and obtaining the executable binary.

### macOS

On macOS, the automated build process `release-macos.sh` is used.
This process fetches and builds the latest IR versions compiled for Windows via the [julec-ir](https://github.com/julelang/julec-ir) repository.
The release archive will be added to the workspace directory where you run the script, named according to the platform–architecture combination (for example, `jule-macos-arm64.tar.xz`).

Dependencies:
- `curl`
- `7zz` (7-Zip)
- `clang++` (Apple Clang, for ARM64 and AMD64)

> [!NOTE]
> For macOS, a native machine or a VM should be used.
>
> It is recommended to build on an ARM64 system.\
> Use Apple Clang for Clang dependency.

> [!TIP]
> On macOS, static linking is not required, as all dependencies should be available internally in a standard macOS installation.

### Windows

On Windows, the automated build process `release-windows.ps1` is used.
This process fetches and builds the latest IR versions compiled for Windows via the [julec-ir](https://github.com/julelang/julec-ir) repository.
The release archive will be added to the workspace directory where you run the script, named according to the platform–architecture combination (for example, `jule-windows-amd64.zip`).

Dependencies:
- `curl.exe` (modern Windows versions have it by default)
- `7z.exe` (7-Zip)
- `clang++.exe` (MinGW-Clang, for ARM64)
- `x86_64-w64-mingw32-clang++.exe` (MinGW-Clang, for AMD64)

> [!NOTE]
> For Windows, a native machine or a VM should be used.
>
> It is recommended to build on an ARM64 system.\
> Use MinGW-Clang for Clang dependency.
> It includes the `clang++.exe` for ARM64, and `x86_64-w64-mingw32-clang++.exe` for AMD64 by default.

### Linux (devcontainers)

On Linux, the automated build process located under `.devcontainers/release` is used.
This process fetches and builds the latest IR versions compiled for Linux via the [julec-ir](https://github.com/julelang/julec-ir) repository.
The release archive will be added to the directory where you run the script, named according to the platform–architecture combination (for example, `jule-linux-amd64.tar.xz`).

It is recommended to run the script from the root directory of the source code.

Like this: `sh ./.devcontainers/release/build.sh`

## Release

- Create a new release on GitHub.
- Create a new tag for the corresponding version, for example `jule0.2.0`.
- Add the release notes and attach the archives you prepared for each platform–architecture combination as release assets.
- Before publishing the release, publish the Manual release for the corresponding version.
- Before publishing the release, publish the IR release for the corresponding version.
- Publish the release.

## Post Release

- The `Version` constant in the `std/jule` package must be set to the new planned version with `@master` tag, like `jule0.2.1 @master`.
- The Manual is always updated against the `@master` branch. Therefore, a link to the released version should be added to the list of older versions in the [Foreword](https://manual.jule.dev/foreword) section.
- Update the `latest` and `current` options to the latest version of [setup-jule](https://github.com/julelang/setup-jule).
