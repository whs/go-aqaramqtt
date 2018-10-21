# Xiaomi Aqara MQTT

I've been using a custom Xiaomi Aqara MQTT bridge for a while, but it loads my Raspberry Pi Model B a lot as it was written in Python.

This project includes

- [Xiaomi Aqara Protocol Client](aqara)
- Coming soon: MQTT bridge

## Building

This project use [Bazel](https://bazel.build) to build, although `go build` would probably works fine if you have the dependencies installed.

To build and run:

```sh
bazel run :go-aqaramqtt
```

Bazel will download and install all dependencies, including the Go compiler!

For cross compiling to Raspberry Pi:

```sh
bazel build :rpi
```

The output binary will be in `bazel-bin/linux_arm_pure_stripped/rpi`

To update dependencies using [bazel-gazelle](https://github.com/bazelbuild/bazel-gazelle). This should be run when you've created new files or import an external module.

```sh
bazel run :gazelle
```

To add external dependencies:

```sh
bazel run :gazelle -- update-repos github.com/example/module
```
