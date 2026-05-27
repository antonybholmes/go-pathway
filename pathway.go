package genes

import (
	"database/sql"
	"math"
	"strings"

	"github.com/antonybholmes/go-basemath"
	"github.com/antonybholmes/go-sys"
	"github.com/antonybholmes/go-sys/db"
	"github.com/antonybholmes/go-sys/log"
)

const (
	// To match MSigDB though unclear where they got this number
	GenesInUniverse = 42577 //45956

	DatasetsSql = `SELECT DISTINCT
		d.public_id,
		d.name,
		c.public_id,
		c.name,
		COUNT(p.id)
		FROM datasets d
		JOIN collections c ON c.dataset_id = d.id
		JOIN pathways p ON p.collection_id = c.id
		GROUP BY c.name 
		ORDER BY d.name, c.name`

	AllPathwaysSql = `SELECT DISTINCT 
		pathway.organization, 
		pathway.dataset, 
		pathway.name, 
		pathway.gene_count, 
		pathway.genes 
		FROM pathway 
		ORDER BY pathway.organization, pathway.dataset, pathway.name`

	CollectionSql = `SELECT 
		c.id, 
		c.public_id, 
		c.name, 
		p.id,
		p.public_id,
		p.name, 
		g.name
		FROM collections c
		JOIN pathways p ON p.collection_id = c.id
		JOIN pathway_genes pg ON pg.pathway_id = p.id
		JOIN genes g ON g.id = pg.gene_id 
		WHERE c.id = :id
		ORDER BY c.name, p.name, g.gene_symbol`

	DatasetCollectionsSql = `SELECT
		d.id,
		d.public_id,
		d.name,
		c.id, 
		c.public_id, 
		c.name, 
		p.id,
		p.public_id,
		p.name, 
		g.name
		FROM datasets d
		JOIN collections c ON c.dataset_id = d.id
		JOIN pathways p ON p.collection_id = c.id
		JOIN pathway_genes pg ON pg.pathway_id = p.id
		JOIN genes g ON g.id = pg.gene_id 
		WHERE d.id IN <<IN>>
		ORDER BY d.name,c.name, p.name, g.gene_symbol`

	GenesSql = `SELECT DISTINCT genes.gene_symbol FROM genes ORDER BY genes.gene_symbol`
)

type (
	Pathway = struct {
		db.Entity
		Genes []string `json:"genes"`
	}

	// Pathway = struct {
	// 	Id    string           `json:"id"`
	// 	Genes *sys.Set[string] `json:"genes"`
	// 	Name  string           `json:"name"`
	// }

	// Geneset struct {
	// 	db.Entity
	// 	Genes []string `json:"genes"`
	// }

	DatasetInfo struct {
		db.Entity
		Collections []*CollectionInfo `json:"collections"`
	}

	CollectionInfo struct {
		db.Entity
		Count int `json:"pathways"`
	}

	Collection struct {
		db.Entity
		Pathways []*Pathway `json:"pathways"`
	}

	Dataset struct {
		db.Entity
		Pathways []*Pathway `json:"pathways"`
	}

	PathwayDB struct {
		db   *sql.DB
		file string
	}

	PathwayOverlaps struct {
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
)

// func (geneset Pathway) ToPathway() *Pathway {
// 	p := NewPathway(sys.NanoId(), geneset.Name, geneset.Genes)

// 	return p
// }

func NewDatasetInfo(id int, publicId string, name string) *DatasetInfo {
	var d DatasetInfo

	d.Id = id
	d.PublicId = publicId
	d.Name = name
	d.Collections = make([]*CollectionInfo, 0, 100)

	return &d
}

func NewCollectionInfo(id int, publicId string, name string, count int) *CollectionInfo {
	var c CollectionInfo

	c.Id = id
	c.PublicId = publicId
	c.Name = name
	c.Count = count

	return &c
}

func NewDataset(id int, publicId string, name string) *Dataset {
	var d Dataset

	d.Id = id
	d.PublicId = publicId
	d.Name = name
	d.Pathways = make([]*Pathway, 0, 100)

	return &d
}

func NewCollection(id int, publicId string, name string) *Collection {
	var c Collection

	c.Id = id
	c.PublicId = publicId
	c.Name = name
	c.Pathways = make([]*Pathway, 0, 100)

	return &c
}

func NewPathway(id int, publicId string, name string) *Pathway {
	var p Pathway

	p.Id = id
	p.PublicId = publicId
	p.Name = name
	p.Genes = make([]string, 0, 100)

	return &p
}

// func (cache *PathwayDBCache) Close() {
// 	for _, db := range cache.cacheMap {
// 		db.Close()
// 	}
// }

func NewPathwayDB(file string) *PathwayDB {

	db := sys.Must(sql.Open(db.Sqlite3DB, file))

	// defer db.Close()

	// genes := sys.NewStringSet()

	// rows := sys.Must(db.Query(GenesSql))

	// defer rows.Close()

	// var gene string

	// for rows.Next() {

	// 	err := rows.Scan(&gene)

	// 	if err != nil {
	// 		log.Fatal().Msgf("cannot read genes")
	// 	}

	// 	genes.Add(gene)
	// }

	// log.Debug().Msgf("Pathway genes: %s %d", file, genes.Len())

	return &PathwayDB{file: file, db: db}
}

func (pdb *PathwayDB) Close() error {
	return pdb.db.Close()
}

func (pdb *PathwayDB) GenesList() ([]string, error) {
	genes := make([]string, 0, 20000)

	rows, err := pdb.db.Query(GenesSql)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var gene string

	for rows.Next() {

		err := rows.Scan(&gene)

		if err != nil {
			return nil, err
		}

		genes = append(genes, gene)
	}

	return genes, nil //pathwaydb.genes.Keys()
}

func (pdb *PathwayDB) Genes() (*sys.StringSet, error) {
	genes := sys.NewStringSet()

	rows, err := pdb.db.Query(GenesSql)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var gene string

	for rows.Next() {

		err := rows.Scan(&gene)

		if err != nil {
			return nil, err
		}

		genes.Add(gene)
	}

	return genes, nil //pathwaydb.genes.Keys()
}

func (pdb *PathwayDB) AllDatasetsInfo() ([]*DatasetInfo, error) {

	rows, err := pdb.db.Query(DatasetsSql)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ret := make([]*DatasetInfo, 0, 10)

	var datasetId int
	var datasetPublicId string
	var dataset string
	var collectionId int
	var collectionPublicId string
	var collection string
	var count int // pathway count in collection
	currentDataset := ""
	datasetIndex := -1

	for rows.Next() {
		err := rows.Scan(&datasetId, &datasetPublicId, &dataset, &collectionId, &collectionPublicId, &collection, &count)

		if err != nil {
			return nil, err
		}

		if dataset != currentDataset {
			ret = append(ret, NewDatasetInfo(datasetId, datasetPublicId, dataset))
			currentDataset = dataset
			datasetIndex++
		}

		cs := NewCollectionInfo(collectionId, collectionPublicId, collection, count)

		ret[datasetIndex].Collections = append(ret[datasetIndex].Collections, cs)

	}

	return ret, nil
}

func (pdb *PathwayDB) GetCollection(id string) (*Collection, error) {

	// log.Debug().Msgf("%v", fmt.Sprintf("'%s'", strings.Join(datasets, "','")))

	// args := make([]interface{}, len(datasets))
	// inRHS := make([]string, len(datasets))

	// for i := range inRHS {
	// 	args[i] = datasets[i]
	// 	inRHS[i] = "?"
	// }

	rows, err := pdb.db.Query(CollectionSql, sql.Named("id", id))

	if err != nil {
		log.Debug().Msgf("e2 %s", err)
		return nil, err
	}

	defer rows.Close()

	var collection *Collection = nil

	var collectionId int
	var collectionPublicId string
	var collectionName string
	var pathwayId int
	var pathwayPublicId string
	var pathwayName string
	var gene string

	for rows.Next() {

		//gene.Taxonomy = tax

		err := rows.Scan(
			&collectionId,
			&collectionPublicId,
			&collectionName,
			&pathwayId,
			&pathwayPublicId,
			&pathwayName,
			&gene)

		if err != nil {
			return nil, err
		}

		if collection == nil {
			collection = NewCollection(collectionId, collectionPublicId, collectionName)
		}

		pathway := NewPathway(pathwayId, pathwayPublicId, pathwayName)

		pathway.Genes = append(pathway.Genes, gene)

		collection.Pathways = append(collection.Pathways, pathway)
	}

	//sql := strings.Replace(PATHWAY_SQL, "<in>", strings.Join(inRHS, ","), 1)

	//log.Debug().Msgf("%v %v", sql, args)

	//log.Debug().Msgf("%v", ret)

	return collection, nil
}

// func (pdb *PathwayDB) MakeDataset(org string, name string) (*Collection, error) {

// 	// log.Debug().Msgf("%v", fmt.Sprintf("'%s'", strings.Join(datasets, "','")))

// 	// args := make([]interface{}, len(datasets))
// 	// inRHS := make([]string, len(datasets))

// 	// for i := range inRHS {
// 	// 	args[i] = datasets[i]
// 	// 	inRHS[i] = "?"
// 	// }

// 	rows, err := pdb.db.Query(PathwaysSql, sql.Named("org", org), sql.Named("dataset", name))

// 	if err != nil {
// 		log.Debug().Msgf("e2 %s", err)
// 		return nil, err
// 	}

// 	defer rows.Close()

// 	dataset := NewDataset(org, name)

// 	var publicId string
// 	var genes string
// 	var geneCount int

// 	for rows.Next() {

// 		//gene.Taxonomy = tax

// 		err := rows.Scan(
// 			&publicId,
// 			&name,
// 			&geneCount,
// 			&genes)

// 		if err != nil {
// 			return nil, err
// 		}

// 		pathway := NewPathway(publicId, name, strings.Split(genes, ","))

// 		dataset.Pathways = append(dataset.Pathways, pathway)
// 	}

// 	//sql := strings.Replace(PATHWAY_SQL, "<in>", strings.Join(inRHS, ","), 1)

// 	//log.Debug().Msgf("%v %v", sql, args)

// 	//log.Debug().Msgf("%v", ret)

// 	return dataset, nil
// }

// Given the names of datasets, produce objects containing all the
// pathways of those datasets
func (pdb *PathwayDB) GetDatasetCollections(datasets []string) ([]*Collection, error) {

	ret := make([]*Collection, 0, len(datasets))

	sql := DatasetCollectionsSql

	args := make([]any, len(datasets))

	sql = db.MakeInSql(sql, "<<IN>>", datasets, &args)

	rows, err := pdb.db.Query(sql, args...)

	if err != nil {
		log.Debug().Msgf("e2 %s", err)
		return nil, err
	}

	defer rows.Close()

	var collection *Collection = nil

	var datasetId int
	var datasetPublicId string
	var datasetName string
	var collectionId int
	var collectionPublicId string
	var collectionName string
	var pathwayId int
	var pathwayPublicId string
	var pathwayName string
	var gene string

	for rows.Next() {

		//gene.Taxonomy = tax

		err := rows.Scan(
			&datasetId,
			&datasetPublicId,
			&datasetName,
			&collectionId,
			&collectionPublicId,
			&collectionName,
			&pathwayId,
			&pathwayPublicId,
			&pathwayName,
			&gene)

		if err != nil {
			return nil, err
		}

		if collection == nil || collection.Id != collectionId {
			collection = NewCollection(collectionId, collectionPublicId, collectionName)
			ret = append(ret, collection)
		}

		pathway := NewPathway(pathwayId, pathwayPublicId, pathwayName)

		pathway.Genes = append(pathway.Genes, gene)

		collection.Pathways = append(collection.Pathways, pathway)
	}

	return ret, nil
}

func (pdb *PathwayDB) NewPathwayOverlaps(geneset *Pathway, collections []*Collection) (*PathwayOverlaps, error) {
	numTests := 0

	collectionNames := make([]string, 0, len(collections))

	genes := sys.NewStringSet()

	for _, collection := range collections {
		numTests += len(collection.Pathways)
		collectionNames = append(collectionNames, collection.Name)

		// all the genes in the datasets we are interested in
		for _, pathway := range collection.Pathways {
			genes.ListUpdate(pathway.Genes)
		}
	}

	// universe of genes
	genes, err := pdb.Genes()

	if err != nil {
		return nil, err
	}

	// see which genes in our test pathway we can use
	validGenes := sys.NewStringSet()

	for _, gene := range geneset.Genes {
		// if genes.Has(gene) {
		// 	usableGenes.Add(gene)
		// }

		// use the universe of all genes to establish is gene is valid or not
		if genes.Has(gene) {
			validGenes.Add(gene)
		}
	}

	ret := PathwayOverlaps{
		Geneset:    geneset.Name,
		Datasets:   collectionNames,
		DatasetIdx: make([]int, numTests),
		Pathway:    make([]string, numTests),
		//ValidGeneCount:       len(*usableGenes),
		GenesInUniverseCount: GenesInUniverse,
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

	return &ret, nil
}

func (pdb *PathwayDB) Overlap(geneset *Pathway, collections []*Collection) (*PathwayOverlaps, error) {
	ret, err := pdb.NewPathwayOverlaps(geneset, collections)

	if err != nil {
		return nil, err
	}

	n := ret.ValidGenes.Len()

	gi := 0
	for ci, collection := range collections {
		for _, pathway := range collection.Pathways {
			K := len(pathway.Genes)

			overlappingGenesInPathway := ret.ValidGenes.Intersect(sys.NewStringSet().ListUpdate(pathway.Genes))

			//overlappingGenes := make([]string, 0, overlappingGenesInPathway.Len())

			//overlappingGenes = append(overlappingGenes, overlappingGenesInPathway.Keys()...)

			// sort overlapping genes for presentation
			//sort.Strings(overlappingGenes)

			k := len(overlappingGenesInPathway.Keys())

			p := float64(1)

			var kDivK float64 = float64(k) / float64(n)

			if k > 0 {
				p = 1 - basemath.HypGeomCDF(k-1, GenesInUniverse, K, n)
			}

			//ret.Name[gi] = geneset.Name
			ret.DatasetIdx[gi] = ci
			ret.Pathway[gi] = pathway.Name
			//ret.TestGenes[gi] = n
			ret.PathwayGeneCounts[gi] = K
			ret.OverlapGeneCounts[gi] = k
			//ret.N[gi] = GENES_IN_UNIVERSE
			ret.PValues[gi] = p
			ret.KDivK[gi] = kDivK
			ret.OverlapGeneList[gi] = strings.Join(overlappingGenesInPathway.Keys(), ",")

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
