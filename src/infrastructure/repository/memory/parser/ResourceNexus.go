package parser

type ResourceNexus struct {
	key     string
	root    bool
	headers []string
	nodes   []ResourceGroup
}
