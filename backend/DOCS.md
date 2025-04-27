To test microservice comms:

```bash
make monolith &&  MICROSERVICE=1 WORLD_GEN_URL=localhost:50054 ./out/monolith
```

```bash
make world_gen && ./out/world_gen_bin --addr=localhost:50054
```

Note: **Do not prepend `localhost:50054` with https!**

To test monolith comms:

`make monolith &&  MICROSERVICE=1 WORLD_GEN_URL=localhost:50054 ./out/monolith`
