auth_svc: auth.proto
	go build -o ./out ./authentication-service

protos: auth.proto game_engine.proto world_gen.proto frame_gen.proto initiator.proto music.proto

%.proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protos/$(basename $@)/$@


