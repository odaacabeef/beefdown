package sequence

const (

	// markdown
	reCodeBlocks = "(?sm)^```beef(.*?)\n^```"

	// playables
	reMult = `\*([[:digit:]]+)`

	// metadata
	reName     = `([0-9A-Za-z'_-]+)`
	reGroup    = `([0-9A-Za-z_-]+)`
	reChannel  = `([0-9]+)`
	reBPM      = `([0-9]+\.?[0-9]+?)`
	reLoop     = `(true|false)`
	reSync     = `(leader)`
	reDivision = `((4th|8th)-triplet|8th|16th|32nd)`
)
