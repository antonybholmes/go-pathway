package pathwaydbcache

import (
	"sync"

	pathway "github.com/antonybholmes/go-pathway"
)

var (
	instance *pathway.PathwayDB
	once     sync.Once
)

func InitCache(file string) *pathway.PathwayDB {
	once.Do(func() {
		instance = pathway.NewPathwayDB(file)
	})

	return instance
}

func GetInstance() *pathway.PathwayDB {
	return instance
}

func Genes() []string {
	return instance.Genes()
}

func AllDatasetsInfo() ([]*pathway.OrganizationInfo, error) {
	return instance.AllDatasetsInfo()
}

func MakePublicDataset(organization string, name string) (*pathway.PublicDataset, error) {
	return instance.MakePublicDataset(organization, name)
}

func Overlap(testPathway *pathway.Pathway, datasets []string) (*pathway.PathwayOverlaps, error) {
	ds, err := instance.MakeDatasets(datasets)

	if err != nil {
		return nil, err
	}

	return instance.Overlap(testPathway, ds)
}
