package sequence

const (

	// markdown
	reCodeBlocks = "(?sm)^```beef(.*?)\n^```"

	// playables
	reNote  = `\b([abcdefg][b,#]?)([[:digit:]]+):?([[:digit:]])?\b`
	reChord = `\b([ABCDEFG][b,#]?)(m7|M7|7|M|m):?([[:digit:]])?\b`
	reMult  = `\*([[:digit:]]+)`

	// metadata
	reName     = `name:([0-9A-Za-z'_-]+)`
	reGroup    = `group:([0-9A-Za-z_-]+)`
	reChannel  = `ch:([0-9]+)`
	reBPM      = `bpm:([0-9]+\.?[0-9]+?)`
	reLoop     = `loop:(true|false)`
	reSync     = `sync:(leader)`
	reDivision = `div:(8th-triplet|8th|16th|32nd)`
)
