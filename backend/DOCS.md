To test microservice comms:

```bash
make monolith &&  MICROSERVICE=1 WORLD_GEN_URL=localhost:50054 ./out/monolith
```

```bash
make worldgen && ./out/worldgen --addr=localhost:50054
```

Note: **Do not prepend `localhost:50054` with https!**

To test monolith comms:

`make monolith &&  MICROSERVICE=1 WORLD_GEN_URL=localhost:50054 ./out/monolith`


For WebTransport, from https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes, may need to:

```
sysctl -w net.core.rmem_max=7500000
sysctl -w net.core.wmem_max=7500000
```
