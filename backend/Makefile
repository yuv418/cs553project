monolith: auth.proto game_engine.proto world_gen.proto frame_gen.proto initiator.proto music.proto
	go build -tags monolith -o ./out/monolith ./bins

initiator: initiator.proto
	go build -tags initiator -o ./out/initiator ./bins

worldgen: world_gen.proto
	go build -tags worldgen -o ./out/worldgen ./bins

engine: game_engine.proto
	go build -tags engine -o ./out/engine ./bins

auth: auth.proto
	go build -tags auth -o ./out/auth ./bins

protos: auth.proto game_engine.proto world_gen.proto frame_gen.proto initiator.proto music.proto

%.proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protos/$(basename $@)/$@

