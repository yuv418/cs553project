# FlappyGo!

FlappyGo! is an enterprise-ready distribution of the classic game "Flappy Bird". It supports an optional, distributed architecutre for its components. Deployers can choose to deploy FlappyGo! as a monolith, with its constituent components running in a single serverside app (plus client), or as microservices, with different components running in different locations and communicating.

## Server(s)

The `backend` directory contains the game server, which is written in Go.

It supports a variety of deployment patterns, where application binaries can be built via `make`.

## Setup

Install a modern version of [Go](https://go.dev/doc/install) 1.24 and [libprotoc](https://protobuf.dev/installation/) 30. 

NOTE: For WebTransport, from https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes, you may need to run:

```
sysctl -w net.core.rmem_max=7500000
sysctl -w net.core.wmem_max=7500000
```

### Monolithic Deployment

Deploying FlappyGo! monolithically can be accomplished by executing:

`make monolith` to build the game.

`MICROSERVICE=0 AUTH_URL=localhost:50051 WORLD_GEN_URL=localhost:50051 INITIATOR_URL=localhost:50051 GAME_ENGINE_URL=localhost:50051 WORLD_GEN_URL=localhost:50051 MUSIC_URL=localhost:50051 SCORE_URL=localhost:50051 ./out/monolith` to run the game.

### Microservice-based Deployment

Deploying FlappyGo! as microservices can be accomplished by executing:

`make <component>` where `<component>` is one of `initiator`, `worldgen`, `engine`, `auth`, `music`, or `score`.

Then, run `./out/<component> --addr=localhost:50054`, again replacing `<component>` with the desired component.

NOTE: Different components require communication with specific other components. Pass the URLs of those services as environment variables when executing the binaries (see Monolith run command above).

## Client

The `flap-client` directory contains the game client, which is written in TypeScript and accessed through a browser.

### Setup

Run `npm install` in the `flap-client` directory to install dependencies (NodeJS, npm are required).

Then, run `npx buf generate` to ensure the latest protobufs are generated.

Update the `.env` file (or create one based on the `.env.sample` file) to point to the services.

### Browser Preparation

FlappyGo!, by default, uses unsigned certificates. WebTransport is only available in [secure contexts](https://developer.mozilla.org/en-US/docs/Web/Security/Secure_Contexts). You may need to close all instances of your browser and relaunch it using a command like:

`<chromium_based_browser_executable> --origin-to-force-quic-on=localhost:4433 --ignore-certificate-errors-spki-list=ibdElbpy/Cl9ZssVrMvsLeXPIGPBHTv/N6KXObqeuKg=` where `<chromium_based_browser_executable>` is replaced.

to trust the local certificate.

### Running

Use `npm run dev` to start a live development server which auto reloads.

Or, use `npm run build` to create a "production ready" pre-built version. It will be output to the `dist` directory.

## Asset Credit

Select assets (sprites, audio) courtesy [Samuel Custodio](https://github.com/samuelcust/flappy-bird-assets).

## Original Reference Game AI Attribution

We sought to use a simple AI example of "Flappy Bird" as our basic game reference.

An initial `reference_game.html` was generated via OpenAI's API using `o1` model with the following prompt:

> Generate HTML/CSS/JS for a clone of the game Flappy Bird. Ensure complete feature parity with the original Flappy Bird game originally released for mobile devices. Output the list of features first, then write the code. Use basic inline SVGs for temporary graphics.

The model returned the following blurb, and the completed HTML file. Minimal changes were made to the file.

>Features of the Flappy Bird Clone
>
>Side-Scrolling Environment
>A continuous scrolling background that loops seamlessly.
>
>Gravity and Flap Mechanics
>The bird experiences constant downward acceleration and rises briefly upon user input (click or tap).
>
>Pipes with Random Gaps
>Infinite procedurally generated pipes with varying gap positions.
>
>Collision Detection
>Checking for collisions with pipes and the ground.
>
>Score Tracking
>Increments whenever the bird successfully passes between the upper and lower pipes.
>
>Difficulty Progression
>Subtle increase in speed and additional challenges as the score grows.
>
>Animated Ground
>A moving ground surface to simulate forward motion.
>
>Game States
>
>Waiting to start
>
>Ongoing gameplay
>
>Game over with replay option
>
>Sound Effects
>Simple collision sound and flap sound triggers when appropriate.
>
>Tap/Click Controls
>User interaction triggers the bird’s upward flap.
>
>Game Over Screen
>Displays final score and best score, with an option to restart instantly.