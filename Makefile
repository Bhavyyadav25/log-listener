build:
			sudo /usr/local/go/bin/go build -o bin/logListener cmd/log-listener/Start.go

run: build
			sudo ./bin/logListener

rfsh: 
		sudo /usr/local/go/bin/go mod tidy