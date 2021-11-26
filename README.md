# tcp2unix
TCP &lt;==> Unix socket file

# Eg

```sh
# Listen 127.0.0.1:9222 forward from unix:/tmp/chrome-run/.devtools.sock, timeout 5m for connection
./tcp2unix-386 127.0.0.1:9222 unix:/tmp/chrome-run/.devtools.sock 5m

# Listen /tmp/chrome-run/.devtools.sock forward from 127.0.0.1:9222, timeout 5m for connection
./tcp2unix-386 unix:/tmp/chrome-run/.devtools.sock 127.0.0.1:9222 5m
```
