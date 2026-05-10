package constants

import "time"

const (
	// keys
	CacheKeyCharDetail = "character:detail:"
	CacheKeyTTSMessage = "tts:message:"

	// ttl
	CharDetailCacheTTL = 10 * time.Minute
	TTSAudioCacheTTL   = 7 * 24 * time.Hour
)
