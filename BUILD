load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/whs/go-aqaramqtt
gazelle(name = "gazelle")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/whs/go-aqaramqtt",
    visibility = ["//visibility:private"],
    deps = [
        "//aqara:go_default_library",
        "@com_github_eclipse_paho_mqtt_golang//:go_default_library",
    ],
)

go_binary(
    name = "go-aqaramqtt",
    embed = [":go_default_library"],
    msan = "auto",
    race = "on",
    visibility = ["//visibility:public"],
)

go_binary(
    name = "rpi",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
    goos = "linux",
    goarch = "arm",
    pure = "on",
)
