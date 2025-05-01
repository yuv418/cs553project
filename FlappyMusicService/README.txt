#dependencies

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

go get github.com/hajimehoshi/ebiten/v2/audio@v2.8.7
go get google.golang.org/grpc@v1.67.0
go get google.golang.org/protobuf@v1.35.1

##protoc compilation

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/music.proto
