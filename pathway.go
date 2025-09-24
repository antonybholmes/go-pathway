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

// To match MSigDB though unclear where they got this number
const GENES_IN_UNIVERSE = 42577 //45956

const DATASETS_SQL = `SELECT DISTINCT 
	pathway.organization, 
	pathway.dataset, 
	COUNT(pathway.id) 
	FROM pathway 
	GROUP BY pathway.organization, pathway.dataset 
	ORDER BY pathway.organization, pathway.dataset`

//const ORG_INFO_SQL = "SELECT DISTINCT pathway.organization, pathway.dataset FROM pathway ORDER BY pathway.organization, pathway.dataset"

const ALL_PATHWAYS_SQL = `SELECT DISTINCT 
	pathway.organization, 
	pathway.dataset, 
	pathway.name, 
	pathway.gene_count, 
	pathway.genes 
	FROM pathway 
	ORDER BY pathway.organization, pathway.dataset, pathway.name`

// const PATHWAYS_SQL = "SELECT dataset, name, genes FROM pathway WHERE dataset IN (<in>) ORDER BY name"
const PATHWAYS_SQL = `SELECT 
	pathway.public_id, 
	pathway.name, 
	pathway.gene_count, 
	pathway.genes 
	FROM pathway 
	WHERE pathway.organization = ?1 AND pathway.dataset = ?2 ORDER BY pathway.name`

const GENES_SQL = `SELECT genes.gene_symbol FROM genes`

type PublicPathway = struct {
	PublicId string   `json:"publicId"`
	Name     string   `json:"name"`
	Genes    []string `json:"genes"`
}

type Pathway = struct {
	PublicId string         `json:"publicId"`
	Genes    *sys.StringSet `json:"genes"`
	Name     string         `json:"name"`
}

func NewPathway(publicId string, name string, genes []string) *Pathway {

	uniqueGenes := sys.NewStringSet().ListUpdate(genes)

	p := Pathway{
		PublicId: publicId,
		Name:     name,
		Genes:    uniqueGenes, //StringSetSort(uniqueGenes),
	}

	return &p
}

type Geneset struct {
	Name  string   `json:"name"`
	Genes []string `json:"genes"`
}

func (geneset Geneset) ToPathway() *Pathway {
	p := NewPathway(sys.NanoId(), geneset.Name, geneset.Genes)

	return p
}

type DatasetInfo struct {
	Organization string `json:"organization"`
	Name         string `json:"name"`
	PathwayCount int    `json:"pathways"`
}

type Dataset struct {
	Organization string     `json:"organization"`
	Name         string     `json:"name"`
	Pathways     []*Pathway `json:"pathways"`
}

type PublicDataset struct {
	Organization string           `json:"organization"`
	Name         string           `json:"name"`
	Pathways     []*PublicPathway `json:"pathways"`
}

type OrganizationInfo struct {
	Name     string         `json:"name"`
	Datasets []*DatasetInfo `json:"datasets"`
}

func NewOrganizationInfo(name string) *OrganizationInfo {
	p := OrganizationInfo{
		Name:     name,
		Datasets: make([]*DatasetInfo, 0, 100),
	}

	return &p
}

func NewDatasetInfo(org string, name string, count int) *DatasetInfo {
	p := DatasetInfo{
		Organization: org,
		Name:         name,
		PathwayCount: count,
	}

	return &p
}

func NewPublicDataset(org string, name string) *PublicDataset {
	p := PublicDataset{
		Organization: org,
		Name:         name,
		Pathways:     make([]*PublicPathway, 0, 100),
	}

	return &p
}

func NewDataset(org string, name string) *Dataset {
	p := Dataset{
		Organization: org,
		Name:         name,
		Pathways:     make([]*Pathway, 0, 100),
	}

	return &p
}

// func (cache *PathwayDBCache) Close() {
// 	for _, db := range cache.cacheMap {
// 		db.Close()
// 	}
// }

type PathwayDB struct {
	genes *sys.Set[string]
	file  string
}

func NewPathwayDB(file string) *PathwayDB {

	db := sys.Must(sql.Open("sqlite3", file))

	defer db.Close()

	genes := sys.NewSet[string]()

	rows := sys.Must(db.Query(GENES_SQL))

	defer rows.Close()

	var gene string

	for rows.Next() {

		err := rows.Scan(&gene)

		if err != nil {
			log.Fatal().Msgf("cannot read genes")
		}

		genes.Add(gene)
	}

	log.Debug().Msgf("Pathway genes: %s %d", file, genes.Len())

	return &PathwayDB{file: file, genes: genes}
}

func (pathwaydb *PathwayDB) Genes() []string {
	return pathwaydb.genes.Keys()
}

func (pathwaydb *PathwayDB) AllDatasetsInfo() ([]*OrganizationInfo, error) {

	db, err := sql.Open("sqlite3", pathwaydb.file)

	if err != nil {
		return nil, err //fmt.Errorf("there was an error with the database query")
	}

	defer db.Close()

	rows, err := db.Query(DATASETS_SQL)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ret := make([]*OrganizationInfo, 0, 10)

	var org string
	var dataset string
	var pathwayCount int
	currentOrg := ""
	orgIndex := -1

	for rows.Next() {
		err := rows.Scan(&org, &dataset, &pathwayCount)

		if err != nil {
			return nil, err
		}

		if org != currentOrg {
			ret = append(ret, NewOrganizationInfo(org))
			currentOrg = org
			orgIndex++
		}

		ds := NewDatasetInfo(org, dataset, pathwayCount)

		ret[orgIndex].Datasets = append(ret[orgIndex].Datasets, ds)

	}

	return ret, nil
}

func (pathwaydb *PathwayDB) MakePublicDataset(org string, name string) (*PublicDataset, error) {

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

	rows, err := db.Query(PATHWAYS_SQL, org, name)

	if err != nil {
		log.Debug().Msgf("e2 %s", err)
		return nil, err
	}

	defer rows.Close()

	dataset := NewPublicDataset(org, name)

	var publicId string
	var genes string
	var geneCount int

	for rows.Next() {

		//gene.Taxonomy = tax

		err := rows.Scan(
			&publicId,
			&name,
			&geneCount,
			&genes)

		if err != nil {
			return nil, err
		}

		pathway := PublicPathway{PublicId: publicId, Name: name, Genes: strings.Split(genes, ",")}

		dataset.Pathways = append(dataset.Pathways, &pathway)
	}

	//sql := strings.Replace(PATHWAY_SQL, "<in>", strings.Join(inRHS, ","), 1)

	//log.Debug().Msgf("%v %v", sql, args)

	//log.Debug().Msgf("%v", ret)

	return dataset, nil
}

func (pathwaydb *PathwayDB) MakeDataset(org string, name string) (*Dataset, error) {

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

	rows, err := db.Query(PATHWAYS_SQL, org, name)

	if err != nil {
		log.Debug().Msgf("e2 %s", err)
		return nil, err
	}

	defer rows.Close()

	dataset := NewDataset(org, name)

	var publicId string
	var genes string
	var geneCount int

	for rows.Next() {

		//gene.Taxonomy = tax

		err := rows.Scan(
			&publicId,
			&name,
			&geneCount,
			&genes)

		if err != nil {
			return nil, err
		}

		pathway := NewPathway(publicId, name, strings.Split(genes, ","))

		dataset.Pathways = append(dataset.Pathways, pathway)
	}

	//sql := strings.Replace(PATHWAY_SQL, "<in>", strings.Join(inRHS, ","), 1)

	//log.Debug().Msgf("%v %v", sql, args)

	//log.Debug().Msgf("%v", ret)

	return dataset, nil
}

// Given the names of datasets, produce objects containing all the
// pathways of those datasets
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
		tokens := strings.SplitN(ds, ":", 2)

		org := tokens[0]
		name := tokens[1]

		dataset, err = pathwaydb.MakeDataset(org, name)

		if err != nil {
			log.Debug().Msgf("e %s", err)
			return nil, err
		}

		ret = append(ret, dataset)
	}

	//sql := strings.Replace(PATHWAY_SQL, "<in>", strings.Join(inRHS, ","), 1)

	//log.Debug().Msgf("%v %v", sql, args)

	//log.Debug().Msgf("%v", ret)

	return ret, nil
}

type PathwayOverlaps struct {
	ValidGenes        *sys.StringSet `json:"-"`
	ValidGeneList     []string       `json:"validGenes"`
	Genes             *sys.StringSet `json:"-"`
	Geneset           string         `json:"geneset"`
	PathwayGeneCounts []int          `json:"pathwayGeneCounts"`
	Pathway           []string       `json:"pathway"`
	OverlapGeneCounts []int          `json:"overlapGeneCounts"`
	KDivK             []float64      `json:"kdivK"`
	PValues           []float64      `json:"pvalues"`
	QValues           []float64      `json:"qvalues"`
	OverlapGeneList   []string       `json:"overlapGeneList"`
	DatasetIdx        []int          `json:"datasetIdx"`
	Datasets          []string       `json:"datasets"`
	//ValidGeneCount       int              `json:"-"`
	GenesInUniverseCount int `json:"genesInUniverseCount"`
}

func (pathwaydb *PathwayDB) NewPathwayOverlaps(geneset *Pathway, datasets []*Dataset) *PathwayOverlaps {
	numTests := 0

	datasetNames := make([]string, 0, len(datasets))

	genes := sys.NewStringSet()

	for _, dataset := range datasets {
		numTests += len(dataset.Pathways)
		datasetNames = append(datasetNames, dataset.Name)

		// all the genes in the datasets we are interested in
		for _, pathway := range dataset.Pathways {
			genes.Update(pathway.Genes)
		}
	}

	// see which genes in our test pathway we can use
	validGenes := sys.NewStringSet()

	for _, gene := range geneset.Genes.Keys() {
		// if genes.Has(gene) {
		// 	usableGenes.Add(gene)
		// }

		// use the universe of all genes to establish is gene is valid or not
		if pathwaydb.genes.Has(gene) {
			validGenes.Add(gene)
		}
	}

	ret := PathwayOverlaps{
		Geneset:    geneset.Name,
		Datasets:   datasetNames,
		DatasetIdx: make([]int, numTests),
		Pathway:    make([]string, numTests),
		//ValidGeneCount:       len(*usableGenes),
		GenesInUniverseCount: GENES_IN_UNIVERSE,
		PathwayGeneCounts:    make([]int, numTests),
		OverlapGeneCounts:    make([]int, numTests),
		KDivK:                make([]float64, numTests),
		PValues:              make([]float64, numTests),
		QValues:              make([]float64, numTests),
		//Log10Q:          make([]float64, n),
		OverlapGeneList: make([]string, numTests),
		Genes:           genes,
		ValidGenes:      validGenes,
		ValidGeneList:   validGenes.Keys(),
	}

	return &ret
}

func (pathwaydb *PathwayDB) Overlap(geneset *Pathway, datasets []*Dataset) (*PathwayOverlaps, error) {
	ret := pathwaydb.NewPathwayOverlaps(geneset, datasets)

	n := ret.ValidGenes.Len()

	gi := 0
	for di, dataset := range datasets {
		for _, pathway := range dataset.Pathways {
			K := pathway.Genes.Len()

			overlappingGenesInPathway := ret.ValidGenes.Intersect(pathway.Genes)

			overlappingGenes := make([]string, 0, overlappingGenesInPathway.Len())

			for _, k := range overlappingGenesInPathway.Keys() {
				overlappingGenes = append(overlappingGenes, k)
			}

			// sort overlapping genes for presentation
			sort.Strings(overlappingGenes)

			k := len(overlappingGenes)

			p := float64(1)

			var kDivK float64 = float64(k) / float64(n)

			if k > 0 {
				p = 1 - basemath.HypGeomCDF(k-1, GENES_IN_UNIVERSE, K, n)
			}

			//ret.Name[gi] = geneset.Name
			ret.DatasetIdx[gi] = di
			ret.Pathway[gi] = pathway.Name
			//ret.TestGenes[gi] = n
			ret.PathwayGeneCounts[gi] = K
			ret.OverlapGeneCounts[gi] = k
			//ret.N[gi] = GENES_IN_UNIVERSE
			ret.PValues[gi] = p
			ret.KDivK[gi] = kDivK
			ret.OverlapGeneList[gi] = strings.Join(overlappingGenes, ",")

			gi++
		}
	}

	// fdr
	idx := sys.Argsort(ret.PValues)

	qn := float64(len(idx))

	orderedIdx := idx[0]
	ret.QValues[orderedIdx] = math.Min(1, math.Max(0, ret.PValues[orderedIdx]*qn))

	for c := 1; c < len(idx); c++ {
		orderedIdx := idx[c]
		rank := float64(c + 1)
		var q float64 = (ret.PValues[orderedIdx] * qn) / rank

		ret.QValues[orderedIdx] = math.Min(
			1,
			math.Max(0, math.Max(ret.QValues[idx[c-1]], q)),
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
