package genes

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/antonybholmes/go-basemath"
	"github.com/antonybholmes/go-sys"
	"github.com/rs/zerolog/log"
)

const GENES_IN_UNIVERSE = 45956

const DATASET_SQL = "SELECT DISTINCT dataset FROM pathway ORDER BY dataset"

const PATHWAY_SQL = "SELECT dataset, name, genes FROM pathway WHERE dataset IN (<in>) ORDER BY name"

type Pathway = struct {
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

type Geneset struct {
	Name  string   `json:"name"`
	Genes []string `json:"genes"`
}

func (geneset Geneset) ToPathway() *Pathway {
	p := NewPathway(geneset.Name)
	p.Genes.UpdateList(geneset.Genes)

	return p
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

func (pathwaydb *PathwayDB) Datasets() (*[]string, error) {

	db, err := sql.Open("sqlite3", pathwaydb.file)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	defer db.Close()

	rows, err := db.Query(DATASET_SQL)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ret := make([]string, 0, 5)

	for rows.Next() {
		var dataset string

		err := rows.Scan(&dataset)

		if err != nil {
			return nil, err
		}

		ret = append(ret, dataset)
	}

	return &ret, nil
}

func (pathwaydb *PathwayDB) MakePathwayCollection(datasets []string) (*PathwayCollection, error) {

	db, err := sql.Open("sqlite3", pathwaydb.file)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	defer db.Close()

	log.Debug().Msgf("%v", fmt.Sprintf("'%s'", strings.Join(datasets, "','")))

	args := make([]interface{}, len(datasets))
	inRHS := make([]string, len(datasets))

	for i := range inRHS {
		args[i] = datasets[i]
		inRHS[i] = "?"
	}

	sql := strings.Replace(PATHWAY_SQL, "<in>", strings.Join(inRHS, ","), 1)

	log.Debug().Msgf("%v %v", sql, args)

	rows, err := db.Query(sql, args...)

	if err != nil {
		log.Debug().Msgf("e %s", err)
		return nil, err
	}

	defer rows.Close()

	var ret PathwayCollection
	var name string
	var dataset string
	var genes string
	ret.Genesets = make([]*Pathway, 0, 5)

	log.Debug().Msgf("herer")

	for rows.Next() {

		//gene.Taxonomy = tax

		err := rows.Scan(
			&dataset,
			&name,
			&genes)

		if err != nil {
			return nil, err
		}

		pathway := NewPathway(name)

		for _, gene := range strings.Split(genes, ",") {
			pathway.Genes.Add(gene)
		}

		ret.Genesets = append(ret.Genesets, pathway)
	}

	ret.Name = dataset

	//log.Debug().Msgf("%v", ret)

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

	log.Debug().Msgf("pathway step 1 %d", n)

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

	log.Debug().Msgf("pathway step 2")

	// fdr
	idx := sys.Argsort(ret.P)

	log.Debug().Msgf("pathway step 3 %d", len(ret.P))

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

	log.Debug().Msgf("pathway step 4")

	for c := range idx {
		ret.Log10Q[c] = -math.Log10(ret.Q[c])
	}

	return ret, nil

}
