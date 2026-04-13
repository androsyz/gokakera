package gokakera

// Progress carries upload progress information delivered to a ProgressFunc callback
// after each successfully stored chunk.
type Progress struct {
	SessionID      string
	ChunkIndex     int
	ChunksReceived int
	TotalChunks    int
	BytesReceived  int64
	TotalSize      int64
	Percentage     float64
}

// ProgressFunc is the callback type invoked after each chunk is successfully uploaded.
type ProgressFunc func(Progress)
