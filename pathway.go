package genes

import (
	"database/sql"
	"sort"
	"strings"

	"os"
	"path/filepath"

	"github.com/antonybholmes/go-basemath"
	"github.com/antonybholmes/go-sys"
	"github.com/rs/zerolog/log"
)

const GENES_IN_UNIVERSE = 45956

const PATHWAY_SQL = "SELECT db, name, genes FROM pathways ORDER BY name"

type Pathway struct {
	Name  string          `json:"name"`
	Genes sys.Set[string] `json:"genes"`
}

type PathwayCollection struct {
	Name     string     `json:"name"`
	Genesets []*Pathway `json:"pathways"`
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

func (pathwaydb *PathwayDB) Pathways() (*PathwayCollection, error) {

	db, err := sql.Open("sqlite3", pathwaydb.file)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	defer db.Close()

	rows, err := db.Query(PATHWAY_SQL)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ret PathwayCollection
	var database string
	var genes string
	ret.Genesets = make([]*Pathway, 0, 5)

	for rows.Next() {
		var pathway Pathway

		//gene.Taxonomy = tax

		err := rows.Scan(
			&database,
			&pathway.Name,
			&genes)

		if err != nil {
			return nil, err
		}

		for _, gene := range strings.Split(genes, ",") {
			pathway.Genes.Add(gene)
		}

		ret.Genesets = append(ret.Genesets, &pathway)
	}

	ret.Name = database

	return &ret, nil
}

type PathwayTestResult struct {
	Name            string   `json:"name"`
	Collection      string   `json:"collection"`
	Geneset         string   `json:"geneset"`
	TestGenes       uint64   `json:"testGenes"`
	PathwayGenes    uint64   `json:"pathwayGenes"`
	OverlapGenes    uint64   `json:"overlapGenes"`
	N               uint64   `json:"n"`
	KDivN           float64  `json:"kdivn"`
	P               float64  `json:"p"`
	Q               float64  `json:"q"`
	Pathways        uint64   `json:"pathways"`
	OverlapGeneList []string `json:"overlapGeneList"`
}

func Test(testPathway *Pathway, pathways *PathwayCollection) ([]*PathwayTestResult, error) {

	n := uint64(len(testPathway.Genes))

	ret := make([]*PathwayTestResult, 0, 100)

	for _, geneset := range pathways.Genesets {
		K := uint64(len(geneset.Genes))

		overlappingGeneSet := testPathway.Genes.Intersect(geneset.Genes)

		overlappingGenes := make([]string, 0, len(overlappingGeneSet))

		for k := range overlappingGeneSet {
			overlappingGenes = append(overlappingGenes, k)
		}

		// sort overlapping genes for presentation
		sort.Strings(overlappingGenes)

		k := uint64(len(overlappingGenes))

		p := float64(1)

		var kDivN float64 = float64(k) / float64(n)

		if k > 0 {

			p = 1 - basemath.HypGeomCDF(k-1, GENES_IN_UNIVERSE, K, n)
		}

		result := PathwayTestResult{Name: testPathway.Name,
			Collection:      pathways.Name,
			Geneset:         geneset.Name,
			TestGenes:       n,
			PathwayGenes:    K,
			OverlapGenes:    k,
			N:               GENES_IN_UNIVERSE,
			P:               p,
			KDivN:           kDivN,
			OverlapGeneList: overlappingGenes}

		ret = append(ret, &result)

	}

	return ret, nil

}
