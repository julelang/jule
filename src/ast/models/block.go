package models

// Block is code block.
type Block struct {
	IsUnsafe bool
	Deferred bool
	Parent   *Block
	SubIndex int // Index of statement in parent block
	Tree     []Statement
	Func     *Fn

	// If block is the root block, has all labels and gotos of all sub blocks.
	Gotos  *Gotos
	Labels *Labels
}
