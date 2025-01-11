package sequence

const (

	// markdown
	reCodeBlocks = "(?sm)^```beef(.*?)\n^```"

	// playables
	reNote  = `\b([abcdefg][b,#]?)([[:digit:]]+):?([[:digit:]])?\b`
	reChord = `\b([ABCDEFG][b,#]?)(m7|M7|7|M|m):?([[:digit:]])?\b`
	reMult  = `\*([[:digit:]]+)`

	// metadata
	reName     = `([0-9A-Za-z'_-]+)`
	reGroup    = `([0-9A-Za-z_-]+)`
	reChannel  = `([0-9]+)`
	reBPM      = `([0-9]+\.?[0-9]+?)`
	reLoop     = `(true|false)`
	reSync     = `(leader)`
	reDivision = `(8th-triplet|8th|16th|32nd)`
)
