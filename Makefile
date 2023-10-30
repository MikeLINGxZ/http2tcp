all:http2tcp_client_win_amd64.exe http2tcp_client_linux_amd64 http2tcp_client_darwin_arm64 http2tcp_server_win_amd64.exe http2tcp_server_linux_amd64 http2tcp_server_darwin_arm64
http2tcp_client_win_amd64.exe:
	GOOS=windows GOARCH=amd64 go build -o ./bin/http2tcp_client_win_amd64.exe ./cmd/client/
http2tcp_client_linux_amd64:
	GOOS=linux GOARCH=amd64 go build -o ./bin/http2tcp_client_linux_amd64 ./cmd/client/
http2tcp_client_darwin_arm64:
	GOOS=linux GOARCH=arm64 go build -o ./bin/http2tcp_client_darwin_arm64 ./cmd/client/
http2tcp_server_win_amd64.exe:
	GOOS=windows GOARCH=amd64 go build -o ./bin/http2tcp_server_win_amd64.exe ./cmd/server/
http2tcp_server_linux_amd64:
	GOOS=linux GOARCH=amd64 go build -o ./bin/http2tcp_server_linux_amd64 ./cmd/server/
http2tcp_server_darwin_arm64:
	GOOS=linux GOARCH=arm64 go build -o ./bin/http2tcp_server_darwin_arm64 ./cmd/server/

clean:
	rm -rf ./bin