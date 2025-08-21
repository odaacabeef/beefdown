test:
	go test ./sequence
	go test ./sequence/parsers/metadata
	go test ./sequence/parsers/part

install:
	go install .
