Elementum daemon [![Build Status](https://gitlab.com/elgatito/elementum/badges/master/pipeline.svg?ignore_skipped=true)](https://gitlab.com/elgatito/elementum)
======

Fork of the great [Pulsar daemon](https://github.com/steeve/pulsar) and [Quasar daemon](https://github.com/scakemyer/quasar)

# Easy development environment set up

There is an easy way to prepare local environment for running/compiling Elementum.

1. Clone `libtorrent-go` repository:
```
git clone git@github.com:ElementumOrg/libtorrent-go.git
```
2. Run `make local-env` that will create `local-env` directory will all required components.
```
make local-env
```
3. Modify `LOCAL-ENV` variable in `test_build.sh` to use `local-env` directory:
```
  export LOCAL_ENV=$GOPATH/src/github.com/ElementumOrg/libtorrent-go/local-env/
```

Or you can apply environment variables and compile in any other way:
```
  export LOCAL_ENV=$GOPATH/src/github.com/ElementumOrg/libtorrent-go/local-env/
  export PATH=$PATH:$LOCAL_ENV/bin/
  export PKG_CONFIG_PATH=$LOCAL_ENV/lib/pkgconfig
  export SWIG_LIB=$LOCAL_ENV/share/swig/4.1.1/
```


# How to run

1. Build the [cross-compiler](https://github.com/ElementumOrg/cross-compiler) images,
    or alternatively, pull the cross-compiler images from [Docker Hub](https://hub.docker.com/r/elementumorg/cross-compiler):

    ```
    make pull-all
    ```

    Or for a specific platform:
    ```
    make pull PLATFORM=android-x64
    ```

2. Set GOPATH

    ```
    export GOPATH="~/go"
    ```

3. go get

    ```
    go get -d github.com/elgatito/elementum
    ```

    For Windows support, but required for all builds, you also need:

    ```
    go get github.com/mattn/go-isatty
    ```

4. Build libtorrent-go libraries:

    ```
    make libs
    ```

5. Make specific platforms, or all of them:

    Linux-x64
    ```
    make linux-x64
    ```

    Darwin-x64
    ```
    make darwin-x64
    ```

    Windows
    ```
    make windows-x86
    ```

    All platforms
    ```
    make
    ```

Find memory leaks

To find memory leaks, we can use Valgrind or use Sanitizers.

```
/bin/bash test_build.sh sanitize
```
This will build the binary with enabled sanitizer, so just run it and wait for errors in the console.

```
valgrind --leak-check=full ./elementum -disableBackup
```
This will run the binary with Valgrind. When you cose the process, Valgrind will show statistics.
It is better to use usual binary, without sanitizer, and add backup disable options, as well as disable Kodi library integration. Or it will take a lot of time/CPU at the startup (unless you do need it).

# How to release

Release of a binary part of Elementum means compiling binaries for all platforms and putting them into <https://github.com/elgatito/elementum-binaries> repository (if we run on the git tag).

Release is done with `release.sh` script.