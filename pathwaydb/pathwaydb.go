package pathwaydb

import (
	"sync"

	pathway "github.com/antonybholmes/go-pathway"
	"github.com/antonybholmes/go-sys"
)

var (
	instance *pathway.PathwayDB
	once     sync.Once
)

func InitPathwayDB(file string) *pathway.PathwayDB {
	once.Do(func() {
		instance = pathway.NewPathwayDB(file)
	})

	return instance
}

func GetInstance() *pathway.PathwayDB {
	return instance
}

func GeneList() ([]string, error) {
	return instance.GenesList()
}

func Genes() (*sys.StringSet, error) {
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
