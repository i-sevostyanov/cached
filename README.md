## Probe
Is in-memory cache server with simple text-based protocol.

### Features
* TTL per key
* Dump and Restore from disk
* Thread-safety
* Commands:
    * `set <key> <value> <ttl>`
    * `get <key>`
    * `del <key>`
    * `quit`

### Why?
For fun and profit :-)

### How to use
You need just start a server and connect to it from `nc`, like this:

```bash
nc 127.0.0.1 4000
```

Set value:
```bash
set mykey myvalue 300s
OK
```

Get value:
```bash
get mykey
myvalue
```

Delete value:
```bash
del mykey
OK
```

### ToDo
* Tests and benchmarks
* Metrics
* Dump and restore without reflection
* Build in GitHub Actions


