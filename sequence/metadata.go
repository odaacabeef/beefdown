package sequence

import (
	metaparser "github.com/odaacabeef/beefdown/sequence/parsers/metadata"
)

type sequenceMetadata = metaparser.SequenceMetadata
type partMetadata = metaparser.PartMetadata
type arrangementMetadata = metaparser.ArrangementMetadata
type funcArpeggiateMetadata = metaparser.FuncArpeggiateMetadata

func newSequenceMetadata(raw string) (sequenceMetadata, error) {
	return metaparser.ParseSequenceMetadata(raw)
}

func newPartMetadata(raw string) (partMetadata, error) {
	return metaparser.ParsePartMetadata(raw)
}

func newArrangementMetadata(raw string) (arrangementMetadata, error) {
	return metaparser.ParseArrangementMetadata(raw)
}

func newFuncArpeggiateMetadata(raw string) (funcArpeggiateMetadata, error) {
	return metaparser.ParseFuncArpeggiateMetadata(raw)
}
