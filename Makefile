test:
	go test ./device
	go test ./sequence/parsers/part
	go test ./sequence/syllables

install:
	go install .
