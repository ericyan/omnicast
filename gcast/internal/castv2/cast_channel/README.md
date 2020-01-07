This directory contains the original .proto file(s) from the Chromium
project, which is released under 3-clause BSD license.

To update the files:

```
curl "https://chromium.googlesource.com/chromium/src/+/refs/heads/master/components/cast_channel/proto/cast_channel.proto?format=TEXT" | base64 --decode > cast_channel.proto
```

To generate Go code:

```
protoc --gogo_out=./cast_channel cast_channel.proto
```
