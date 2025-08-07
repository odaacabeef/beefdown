test:
	go test ./device
	go test ./sequence
	go test ./sequence/parsers/part

install:
	go install .
