## Cached
Is an in-memory cache server with a simple text-based protocol.

### Features
* TTL per key (based on BTree index)
* Dump and Restore from disk
* Thread-safety
* Commands:
    * `set <key> <value> <ttl>`
    * `get <key>`
    * `del <key>`
    * `stats`
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

Show stats:
```bash
stats
Hit: 865, Miss: 24, Size: 853
```

### ToDo
* Tests and benchmarks
* Metrics
* Dump and restore without reflection
* Build in GitHub Actions
