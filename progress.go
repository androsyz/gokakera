package gokakera

type Progress struct {
	SessionID      string
	ChunkIndex     int
	ChunksReceived int
	TotalChunks    int
	BytesReceived  int64
	TotalSize      int64
	Percentage     float64
}

type ProgressFunc func(Progress)
