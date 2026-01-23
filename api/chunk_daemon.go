package api

// ChunkNewChunkEvent carries chunk payload in a serialisable form.
// - SubChunks/Biomes can be decoded with bedrock-world-operator's chunk.DiskDecode (SerialisedData).
// - BlockEntities is a concatenation of LittleEndian NBT compounds (same format as gophertunnel nbt encoder).
type ChunkNewChunkEvent struct {
	Dimension     int32    `json:"dimension"`
	ChunkX        int32    `json:"chunk_x"`
	ChunkZ        int32    `json:"chunk_z"`
	SubChunks     [][]byte `json:"sub_chunks"`
	Biomes        []byte   `json:"biomes"`
	BlockEntities []byte   `json:"block_entities"`
	TimeStamp     int64    `json:"timestamp"`
}

type ChunkDaemon interface {
	Name() (name string)
	ReConfig(config map[string]interface{}) (err error)
	Config() (config map[string]interface{})

	RegisterWhenNewChunk(handler func(event *ChunkNewChunkEvent)) (string, error)
	UnregisterWhenNewChunk(listenerID string) bool
}
