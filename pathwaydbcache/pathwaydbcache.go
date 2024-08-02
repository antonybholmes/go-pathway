package pathwaydbcache

import (
	"sync"

	pathway "github.com/antonybholmes/go-pathway"
)

var instance *pathway.PathwayDB
var once sync.Once

func InitCache(file string) *pathway.PathwayDB {
	once.Do(func() {
		instance = pathway.NewPathwayDB(file)
	})

	return instance
}

func GetInstance() *pathway.PathwayDB {
	return instance
}

func Datasets() (*[]*pathway.DatasetInfo, error) {
	return GetInstance().Datasets()
}

func Test(testPathway *pathway.Pathway, datasets []string) (*pathway.PathwayTests, error) {
	ds, err := GetInstance().MakeDatasets(datasets)

	if err != nil {
		return nil, err
	}

	return pathway.Test(testPathway, ds)
}
