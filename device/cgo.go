package device

/*
#cgo darwin LDFLAGS: -L../rust/target/release -lbeefdown -framework CoreMIDI -framework CoreFoundation -framework CoreAudio
#cgo linux LDFLAGS: -L../rust/target/release -lbeefdown -lasound -lpthread -ldl -lm
#cgo windows LDFLAGS: -L../rust/target/release -lbeefdown -lwinmm -lws2_32 -luserenv
*/
import "C"
