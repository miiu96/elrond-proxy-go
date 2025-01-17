package data

// AtlasBlock is a block, as required by BlockAtlas
// Will be removed when using the "hyperblock" route in BlockAtlas as well.
type AtlasBlock struct {
	Nonce        uint64                `form:"nonce" json:"nonce"`
	Hash         string                `form:"hash" json:"hash"`
	Transactions []DatabaseTransaction `form:"transactions" json:"transactions"`
}

// BlockApiResponse is a response holding a block
type BlockApiResponse struct {
	Data  BlockApiResponsePayload `json:"data"`
	Error string                  `json:"error"`
	Code  ReturnCode              `json:"code"`
}

// BlockApiResponsePayload wraps a block
type BlockApiResponsePayload struct {
	Block Block `json:"block"`
}

// Block is a block
type Block struct {
	Nonce           uint64            `json:"nonce"`
	Round           uint64            `json:"round"`
	Hash            string            `json:"hash"`
	PrevBlockHash   string            `json:"prevBlockHash"`
	Epoch           uint32            `json:"epoch"`
	Shard           uint32            `json:"shard"`
	NumTxs          uint32            `json:"numTxs"`
	NotarizedBlocks []*NotarizedBlock `json:"notarizedBlocks,omitempty"`
	MiniBlocks      []*MiniBlock      `json:"miniBlocks,omitempty"`
}

// NotarizedBlock is a notarized block
type NotarizedBlock struct {
	Hash  string `json:"hash"`
	Nonce uint64 `json:"nonce"`
	Shard uint32 `json:"shard"`
}

// MiniBlock is a miniblock
type MiniBlock struct {
	Hash             string             `json:"hash"`
	Type             string             `json:"type"`
	SourceShard      uint32             `json:"sourceShard"`
	DestinationShard uint32             `json:"destinationShard"`
	Transactions     []*FullTransaction `json:"transactions,omitempty"`
}

// HyperblockApiResponse is a response holding a hyperblock
type HyperblockApiResponse struct {
	Data  HyperblockApiResponsePayload `json:"data"`
	Error string                       `json:"error"`
	Code  ReturnCode                   `json:"code"`
}

// NewHyperblockApiResponse creates a HyperblockApiResponse
func NewHyperblockApiResponse(hyperblock Hyperblock) *HyperblockApiResponse {
	return &HyperblockApiResponse{
		Data: HyperblockApiResponsePayload{
			Hyperblock: hyperblock,
		},
		Code: ReturnCodeSuccess,
	}
}

// HyperblockApiResponsePayload wraps a hyperblock
type HyperblockApiResponsePayload struct {
	Hyperblock Hyperblock `json:"hyperblock"`
}

// Hyperblock contains all fully executed (both in source and in destination shards) transactions notarized in a given metablock
type Hyperblock struct {
	Nonce         uint64             `json:"nonce"`
	Round         uint64             `json:"round"`
	Hash          string             `json:"hash"`
	PrevBlockHash string             `json:"prevBlockHash"`
	Epoch         uint32             `json:"epoch"`
	NumTxs        uint32             `json:"numTxs"`
	ShardBlocks   []*NotarizedBlock  `json:"shardBlocks"`
	Transactions  []*FullTransaction `json:"transactions"`
}
