# tcp2unix
TCP &lt;==> Unix socket file

# Eg

```sh
./tcp2unix-386 127.0.0.1:9222 unix:/tmp/chrome-run/.devtools.sock 5m
./tcp2unix-386 unix:/tmp/chrome-run/.devtools.sock 127.0.0.1:9222 5m
```
