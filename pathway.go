package genes

import (
	"database/sql"
	"math"
	"sort"
	"strings"

	"github.com/antonybholmes/go-basemath"
	"github.com/antonybholmes/go-sys"
	"github.com/rs/zerolog/log"
)

const GENES_IN_UNIVERSE = 45956

const DATASET_SQL = "SELECT DISTINCT pathway.dataset, COUNT(pathway.id) FROM pathway GROUP BY pathway.dataset ORDER BY pathway.dataset"

// const PATHWAY_SQL = "SELECT dataset, name, genes FROM pathway WHERE dataset IN (<in>) ORDER BY name"
const PATHWAY_SQL = "SELECT pathway.name, pathway.genes FROM pathway WHERE pathway.dataset = ?1 ORDER BY pathway.name"

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

type DatasetInfo struct {
	Name     string `json:"name"`
	Pathways int    `json:"pathways"`
}

type Dataset struct {
	Name     string     `json:"name"`
	Pathways []*Pathway `json:"pathways"`
}

func NewDataset(name string) *Dataset {
	p := Dataset{
		Name:     name,
		Pathways: make([]*Pathway, 0, 100),
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

func (pathwaydb *PathwayDB) Datasets() (*[]*DatasetInfo, error) {

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

	ret := make([]*DatasetInfo, 0, 5)

	for rows.Next() {
		var dataset DatasetInfo

		err := rows.Scan(&dataset.Name, &dataset.Pathways)

		if err != nil {
			return nil, err
		}

		ret = append(ret, &dataset)
	}

	return &ret, nil
}

func (pathwaydb *PathwayDB) MakeDatasets(datasets []string) ([]*Dataset, error) {

	db, err := sql.Open("sqlite3", pathwaydb.file)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	defer db.Close()

	// log.Debug().Msgf("%v", fmt.Sprintf("'%s'", strings.Join(datasets, "','")))

	// args := make([]interface{}, len(datasets))
	// inRHS := make([]string, len(datasets))

	// for i := range inRHS {
	// 	args[i] = datasets[i]
	// 	inRHS[i] = "?"
	// }

	var dataset *Dataset

	ret := make([]*Dataset, 0, len(datasets))

	for _, ds := range datasets {
		rows, err := db.Query(PATHWAY_SQL, ds)

		if err != nil {
			log.Debug().Msgf("e %s", err)
			return nil, err
		}

		defer rows.Close()

		dataset = NewDataset(ds)

		var name string

		var genes string

		for rows.Next() {

			//gene.Taxonomy = tax

			err := rows.Scan(

				&name,
				&genes)

			if err != nil {
				return nil, err
			}

			pathway := NewPathway(name)

			for _, gene := range strings.Split(genes, ",") {
				pathway.Genes.Add(gene)
			}

			dataset.Pathways = append(dataset.Pathways, pathway)
		}

		ret = append(ret, dataset)
	}

	//sql := strings.Replace(PATHWAY_SQL, "<in>", strings.Join(inRHS, ","), 1)

	//log.Debug().Msgf("%v %v", sql, args)

	//log.Debug().Msgf("%v", ret)

	return ret, nil
}

type PathwayTests struct {
	Geneset      string    `json:"geneset"`
	Datasets     []string  `json:"datasets"`
	DatasetIdx   []int     `json:"datasetIdx"`
	Pathway      []string  `json:"pathway"`
	GenesetSize  int       `json:"n"`
	N            int       `json:"N"`
	PathwayGenes []int     `json:"nPathwayGenes"`
	OverlapGenes []int     `json:"nOverlapGenes"`
	KDivN        []float64 `json:"kdivn"`
	P            []float64 `json:"p"`
	Q            []float64 `json:"q"`
	//Log10Q          []float64 `json:"log10q"`
	OverlapGeneList []string         `json:"overlapGeneList"`
	Genes           *sys.Set[string] `json:"-"`
	UsableGenes     *sys.Set[string] `json:"-"`
}

func NewPathwayTests(geneset *Pathway, datasets []*Dataset) *PathwayTests {
	var numTests = 0

	datasetNames := make([]string, 0, len(datasets))

	genes := sys.NewSet[string]()

	for _, dataset := range datasets {
		numTests += len(dataset.Pathways)
		datasetNames = append(datasetNames, dataset.Name)

		for _, pathway := range dataset.Pathways {
			genes.Update(pathway.Genes)
		}
	}

	usableGenes := sys.NewSet[string]()

	for gene := range *(geneset.Genes) {
		if genes.Has(gene) {
			usableGenes.Add(gene)
		}
	}

	ret := PathwayTests{
		Geneset:      geneset.Name,
		Datasets:     datasetNames,
		DatasetIdx:   make([]int, numTests),
		Pathway:      make([]string, numTests),
		GenesetSize:  len(*usableGenes),
		N:            GENES_IN_UNIVERSE,
		PathwayGenes: make([]int, numTests),
		OverlapGenes: make([]int, numTests),
		KDivN:        make([]float64, numTests),
		P:            make([]float64, numTests),
		Q:            make([]float64, numTests),
		//Log10Q:          make([]float64, n),
		OverlapGeneList: make([]string, numTests),
		Genes:           genes,
		UsableGenes:     usableGenes,
	}

	return &ret
}

func Test(geneset *Pathway, datasets []*Dataset) (*PathwayTests, error) {
	ret := NewPathwayTests(geneset, datasets)

	//n := len(*usableGenes)

	gi := 0
	for di, dataset := range datasets {
		for _, pathway := range dataset.Pathways {
			K := len(*pathway.Genes)

			overlappingGeneSet := ret.UsableGenes.Intersect(pathway.Genes)

			overlappingGenes := make([]string, 0, len(*overlappingGeneSet))

			for k := range *overlappingGeneSet {
				overlappingGenes = append(overlappingGenes, k)
			}

			// sort overlapping genes for presentation
			sort.Strings(overlappingGenes)

			k := len(overlappingGenes)

			p := float64(1)

			var kDivN float64 = float64(k) / float64(ret.GenesetSize)

			if k > 0 {

				p = 1 - basemath.HypGeomCDF(k-1, GENES_IN_UNIVERSE, K, ret.GenesetSize)
			}

			//ret.Name[gi] = geneset.Name
			ret.DatasetIdx[gi] = di
			ret.Pathway[gi] = pathway.Name
			//ret.TestGenes[gi] = n
			ret.PathwayGenes[gi] = K
			ret.OverlapGenes[gi] = k
			//ret.N[gi] = GENES_IN_UNIVERSE
			ret.P[gi] = p
			ret.KDivN[gi] = kDivN
			ret.OverlapGeneList[gi] = strings.Join(overlappingGenes, ",")

			gi++
		}
	}

	// fdr
	idx := sys.Argsort(ret.P)

	qn := float64(len(idx))

	orderedIdx := idx[0]
	ret.Q[orderedIdx] = math.Min(1, math.Max(0, ret.P[orderedIdx]*qn))

	for c := 1; c < len(idx); c++ {
		orderedIdx := idx[c]
		rank := float64(c + 1)
		var q float64 = (ret.P[orderedIdx] * qn) / rank

		ret.Q[orderedIdx] = math.Min(
			1,
			math.Max(0, math.Max(ret.Q[idx[c-1]], q)),
		)

	}

	// for c := range idx {
	// 	if ret.Q[c] > 0 {
	// 		ret.Log10Q[c] = -math.Log10(ret.Q[c])
	// 	} else {
	// 		ret.Log10Q[c] = 1000
	// 	}
	// }

	return ret, nil

}
