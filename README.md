# Xiaomi Aqara MQTT

I've been using a custom Xiaomi Aqara MQTT bridge for a while, but it loads my Raspberry Pi Model B a lot as it was written in Python.

This project includes

- [Xiaomi Aqara Protocol Client](aqara)
- [MQTT bridge](main.go)

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

## Command line options

Option             | Required | Description
-------------------|----------|-------------------------------------
--help             |          | Read help     
--ip               |          | Xiaomi Gateway IP address
--sid              |          | Xiaomi Gateway SID
--key              | Y        | Xiaomi Gateway encryption key ([Tutorial](https://www.domoticz.com/wiki/Xiaomi_Gateway_(Aqara)#Adding_the_Xiaomi_Gateway_to_Domoticz)). Use environment variable `AQARA_KEY` instead. 
--iface            | Y        | Network adapter to use for Xiaomi communication (eg. eth0)
--mqtt-server      | Y        | Protocol and address of MQTT server (eg. tcp://192.168.1.1:1883. Supported scheme: tcp, ssl, ws)
--username         |          | MQTT username. Use environment variable `MQTT_USERNAME` instead.
--password         |          | MQTT password. Use environment variable `MQTT_PASSWORD` instead. 
--prefix           |          | MQTT prefix. Default to "xiaomi"

## License

Licensed under the [MIT license](LICENSE)

This project is [unmaintained](http://unmaintained.tech/). You may use it, but issues and pull requests might be ignored.

