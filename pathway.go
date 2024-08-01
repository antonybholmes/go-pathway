package genes

import (
	"database/sql"
	"math"
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
	Name  string           `json:"name"`
	Genes *sys.Set[string] `json:"genes"`
}

func NewPathway(name string) *Pathway {
	p := Pathway{
		Name:  name,
		Genes: sys.NewSet[string](),
	}

	return &p
}

type PathwayCollection struct {
	Name     string     `json:"name"`
	Genesets []*Pathway `json:"pathways"`
}

func NewPathwayCollection(name string) *PathwayCollection {
	p := PathwayCollection{
		Name:     name,
		Genesets: make([]*Pathway, 0, 100),
	}

	return &p
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

type PathwayTests struct {
	Name            []string  `json:"name"`
	Collection      []string  `json:"collection"`
	Geneset         []string  `json:"geneset"`
	TestGenes       []uint64  `json:"testGenes"`
	PathwayGenes    []uint64  `json:"pathwayGenes"`
	OverlapGenes    []uint64  `json:"overlapGenes"`
	N               []uint64  `json:"n"`
	KDivN           []float64 `json:"kdivn"`
	P               []float64 `json:"p"`
	Q               []float64 `json:"-"`
	Log10Q          []float64 `json:"log10q"`
	OverlapGeneList []string  `json:"overlapGeneList"`
}

func NewPathwayTests(pathways *PathwayCollection) *PathwayTests {
	n := len(pathways.Genesets)
	ret := PathwayTests{
		Name:            make([]string, n),
		Collection:      make([]string, n),
		Geneset:         make([]string, n),
		TestGenes:       make([]uint64, n),
		PathwayGenes:    make([]uint64, n),
		OverlapGenes:    make([]uint64, n),
		N:               make([]uint64, n),
		KDivN:           make([]float64, n),
		P:               make([]float64, n),
		Q:               make([]float64, n),
		Log10Q:          make([]float64, n),
		OverlapGeneList: make([]string, n),
	}

	return &ret
}

func Test(testPathway *Pathway, pathways *PathwayCollection) (*PathwayTests, error) {

	n := uint64(len(*testPathway.Genes))

	ret := NewPathwayTests(pathways)

	for gi, geneset := range pathways.Genesets {
		K := uint64(len(*geneset.Genes))

		overlappingGeneSet := testPathway.Genes.Intersect(geneset.Genes)

		overlappingGenes := make([]string, 0, len(*overlappingGeneSet))

		for k := range *overlappingGeneSet {
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

		ret.Name[gi] = testPathway.Name
		ret.Collection[gi] = pathways.Name
		ret.Geneset[gi] = geneset.Name
		ret.TestGenes[gi] = n
		ret.PathwayGenes[gi] = K
		ret.OverlapGenes[gi] = k
		ret.N[gi] = GENES_IN_UNIVERSE
		ret.P[gi] = p
		ret.KDivN[gi] = kDivN
		ret.OverlapGeneList[gi] = strings.Join(overlappingGenes, ",")

	}

	// fdr
	idx := sys.Argsort(ret.P)

	qn := float64(len(idx))

	ret.Q[0] = math.Min(1, math.Max(0, ret.P[0]*qn))

	for c := 1; c < len(idx); c++ {
		rank := float64(c + 1)
		var q float64 = (ret.P[idx[c]] * qn) / rank

		ret.Q[c] = math.Min(
			1,
			math.Max(0, math.Max(ret.Q[idx[c-1]], q)),
		)

	}

	for c := range idx {
		ret.Log10Q[c] = -math.Log10(ret.Q[c])
	}

	return ret, nil

}
