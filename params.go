package rsql

// Params stores the result of parsing the rsql query
type Params struct {
	Selects []string
	Filters *Node
	Sorts   []*Sort
	Limit   uint
	Offset  uint
	Cursor  string
}
