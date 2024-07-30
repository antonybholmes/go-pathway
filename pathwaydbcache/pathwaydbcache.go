package pathwaydbcache

import (
	"sync"

	pathway "github.com/antonybholmes/go-pathway"
)

var instance *pathway.PathwayDBCache
var once sync.Once

func InitCache(dir string) *pathway.PathwayDBCache {
	once.Do(func() {
		instance = pathway.NewPathwayDBCache(dir)
	})

	return instance
}

func GetInstance() *pathway.PathwayDBCache {
	return instance
}

func Dir() string {
	return instance.Dir()
}

func PathwayDB(assembly string) (*pathway.PathwayDB, error) {
	return instance.PathwayDB(assembly)
}
