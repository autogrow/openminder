default:
	mkdir bin -p
	GOARCH=arm go build -o ./bin/omcli ./cmd/omcli
	GOARCH=arm go build -o ./bin/openminder ./cmd/openminder