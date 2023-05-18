all:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go
	zip iam-enforcer-lambda.zip main
	rm main