package shortener

type Engine interface {
	Shorten(url, path string) string
	GetUrl(path string) (url string)
}

type EngineType int

const (
	EngineTypeInMemory EngineType = iota
	EngineTypeInSQL
	EngineTypeInRedis
)

func NewEngine(eType EngineType) Engine {
	switch eType {
	default:
		return &InMemoryEngine{
			hashToUrl: make(map[string]string),
			urlToHash: make(map[string]string),
		}
	}
}
