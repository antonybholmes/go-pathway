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

func AllDatasetsInfo() ([]*pathway.DatasetInfo, error) {
	return instance.AllDatasetsInfo()
}

func GetDatasets(ids []string) ([]*pathway.Dataset, error) {
	return instance.GetDatasets(ids)
}

func GetCollection(id string) (*pathway.Collection, error) {
	return instance.GetCollection(id)
}

func GetCollections(ids []string) ([]*pathway.Dataset, error) {
	return instance.GetCollections(ids)
}

func GetGeneSets(ids []string) ([]*pathway.Dataset, error) {
	return instance.GetGeneSets(ids)
}

func Overlap(testPathway *pathway.GeneSet, datasetIds []string) (*pathway.PathwayOverlaps, error) {
	datasets, err := instance.GetDatasets(datasetIds)

	if err != nil {
		return nil, err
	}

	collections := make([]*pathway.Collection, 0, len(datasets))

	for _, dataset := range datasets {
		collections = append(collections, dataset.Collections...)
	}

	return instance.Overlap(testPathway, collections)
}
