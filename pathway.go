package genes

import (
	"database/sql"
	"sort"
	"strings"

	"os"
	"path/filepath"

	"github.com/antonybholmes/go-dna"
	"github.com/rs/zerolog/log"
)

const PATHWAY_SQL = "SELECT name, genes FROM pathways ORDER BY name"

type Pathway struct {
	name  string
	genes []string
}

type PathwayCollection struct {
	name     string
	genesets []*Pathway
}

type PathwayDBCache struct {
	dir      string
	cacheMap map[string]*PathwayDB
}

func NewPathwayDBCache(dir string) *PathwayDBCache {
	cacheMap := make(map[string]*PathwayDB)

	files, err := os.ReadDir(dir)

	log.Debug().Msgf("---- genedb ----")

	if err != nil {
		log.Fatal().Msgf("error opening %s", dir)
	}

	log.Debug().Msgf("caching gene databases in %s...", dir)

	for _, file := range files {
		basename := file.Name()

		if strings.HasSuffix(basename, ".db") {

			name := strings.TrimSuffix(basename, filepath.Ext(basename))
			db := NewPathwayDB(filepath.Join(dir, basename))
			cacheMap[name] = db

			log.Debug().Msgf("found gene database %s", name)
		}
	}

	log.Debug().Msgf("---- end ----")

	return &PathwayDBCache{dir, cacheMap}
}

func (cache *PathwayDBCache) Dir() string {
	return cache.dir
}

func (cache *PathwayDBCache) List() []string {

	ids := make([]string, 0, len(cache.cacheMap))

	for id := range cache.cacheMap {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	return ids
}

func (cache *PathwayDBCache) PathwayDB(name string) (*PathwayDB, error) {

	return cache.cacheMap[name], nil
}

// func (cache *PathwayDBCache) Close() {
// 	for _, db := range cache.cacheMap {
// 		db.Close()
// 	}
// }

type PathwayDB struct {
	file string
}

func NewPathwayDB(file string) *PathwayDB {

	return &PathwayDB{file}
}

func (pathwaydb *PathwayDB) Pathways() (PathwayCollection, error) {

	rows, err := pathwaydb.withinGeneStmt.Query(
		mid,
		level,
		location.Chr,
		location.Start,
		location.End)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	return rowsToRecords(location, rows, level)
}

func (genedb *PathwayDB) WithinGenesAndPromoter(location *dna.Location, level Level, pad uint) (*GenomicFeatures, error) {
	mid := (location.Start + location.End) / 2

	// rows, err := genedb.withinGeneAndPromStmt.Query(
	// 	mid,
	// 	level,
	// 	location.Chr,
	// 	pad,
	// 	location.Start,
	// 	pad,
	// 	location.Start,
	// 	pad,
	// 	location.End,
	// 	pad,
	// 	location.End)

	rows, err := genedb.withinGeneAndPromStmt.Query(
		mid,
		level,
		location.Chr,
		location.Start,
		location.End,
		pad)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	return rowsToRecords(location, rows, level)
}

func (genedb *PathwayDB) InExon(location *dna.Location, geneId string) (*GenomicFeatures, error) {
	mid := (location.Start + location.End) / 2

	// rows, err := genedb.inExonStmt.Query(
	// 	mid,
	// 	geneId,
	// 	location.Chr,
	// 	location.Start,
	// 	location.Start,
	// 	location.End,
	// 	location.End)

	rows, err := genedb.inExonStmt.Query(
		mid,
		geneId,
		location.Chr,
		location.Start,
		location.End)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	return rowsToRecords(location, rows, LEVEL_EXON)
}

func (genedb *PathwayDB) ClosestGenes(location *dna.Location, n uint16, level Level) (*GenomicFeatures, error) {
	mid := (location.Start + location.End) / 2

	// rows, err := genedb.closestGeneStmt.Query(mid,
	// 	level,
	// 	location.Chr,
	// 	mid,
	// 	n)

	rows, err := genedb.closestGeneStmt.Query(mid,
		level,
		location.Chr,
		n)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	return rowsToRecords(location, rows, level)
}

func rowsToRecords(location *dna.Location, rows *sql.Rows, level Level) (*GenomicFeatures, error) {
	defer rows.Close()

	var id uint
	var chr string
	var start uint
	var end uint
	var strand string
	var geneId string
	var geneSymbol string
	var d int

	// 10 seems a reasonable guess for the number of features we might see, just
	// to reduce slice reallocation
	var features = make([]*GenomicFeature, 0, 10)

	for rows.Next() {
		err := rows.Scan(&id, &chr, &start, &end, &strand, &geneId, &geneSymbol, &d)

		if err != nil {
			return nil, err //fmt.Errorf("there was an error with the database records")
		}

		location = dna.NewLocation(chr, start, end)

		feature := GenomicFeature{Id: id,
			Location:   location,
			Strand:     strand,
			GeneId:     geneId,
			GeneSymbol: geneSymbol,
			TssDist:    d}

		features = append(features, &feature)
	}

	return &GenomicFeatures{Location: location, Level: level.String(), Features: features}, nil
}
