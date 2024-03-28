# Releasing

This document explains how the Retro AIM Server release process works.

## Overview

Retro AIM Server is built and released to Github using [GoReleaser](https://goreleaser.com/). The release process, which
runs from a local computer (and not a CI/CD process) creates pre-built binaries for several platforms (Windows, MacOS,
Linux).

GoReleaser runs in a Docker container for two important reasons:

- Docker provides a hermetic environment that prevents build contamination from the host environment.
- The Docker image [goreleaser-cross](https://github.com/goreleaser/goreleaser-cross) consolidates all the libraries
  needed for cross-compilation in one place.
    - Given that the application uses cgo, it's much easier to use a pre-built container to build the sqlite library
      than to manually set up a C compiler on Mac/Windows/Linux etc.

### Code Signing Policy

Given the cost and complexity of code signing, this project only distributes unsigned binaries. This means that
MacOS distrusts Retro AIM Server by default and quarantines the application when you open it.
> If you don't want to bypass this security mechanism, you can [build the project yourself](./building) instead.

### Windows Build Obfuscation

Due to cost and complexity, none of the release artifacts are signed. One consequence of this is that Windows Defender
falsely detects [the `.exe` as a virus](https://go.dev/doc/faq#virus) and auto-quarantines the file upon execution.

We get around this by obfuscating the go binary at built time using [garble](https://github.com/burrowers/garble). In
order to accomplish this, the project wraps GoReleaser and garble in a custom Dockerfile `Dockerfile.goreleaser`.

The GoRelease-garble image must be built locally before running the release process.

## Release Procedure

The following is the procedure that builds Retro AIM Server and uploads the build artifacts to a Github release.

1. **Build Custom Docker Image**

   Build the custom GoReleaser Docker image. Ensure that the latest version is set under `GO_RELEASER_CROSS_VERSION` in
   the project Makefile.

    ```sh
    make goreleaser-docker
    ```

2. **Export Github Personal Access Token (PAT)**

   Export a Github Personal Access Token that has `write:packages` permissions for the Retro AIM Server repo.

    ```sh
    export GITHUB_TOKEN=...
    ```

3. **Tag The Build**

   Tag the build using [semantic versioning](https://semver.org/).
    ```shell
    git tag v0.1.0
    git push --tags
    ```

4. **Dry-Run Release**

   Execute a dry-run build to make sure all the moving parts work together. Fix any problems that crop up before
   continuing.

    ```shell
   make release-dry-run
    ```

5. **Release It!**

   Now run the release process. Once its complete, a new [release](https://github.com/mk6i/retro-aim-server/releases)
   should appear in Github with download artifacts attached.