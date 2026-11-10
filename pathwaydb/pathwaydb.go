package pathwaydb

import (
	"sync"

	pathway "github.com/antonybholmes/go-pathway"
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

func Instance() *pathway.PathwayDB {
	return instance
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
