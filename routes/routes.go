package routes

import (
	"errors"

	pathway "github.com/antonybholmes/go-pathway"
	"github.com/antonybholmes/go-pathway/pathwaydb"
	"github.com/antonybholmes/go-web"
	"github.com/gin-gonic/gin"
)

type (
	ReqOverlapParams struct {
		Geneset  *pathway.GeneSet `json:"geneset"`
		Datasets []string         `json:"datasets"`
	}

	ReqDatasetParams struct {
		Organization string `json:"organization"`
		Name         string `json:"name"`
	}
)

func ParseOverlapParamsFromPost(c *gin.Context) (*ReqOverlapParams, error) {

	var params ReqOverlapParams

	err := c.Bind(&params)

	if err != nil {
		return nil, err
	}

	return &params, nil
}

func ParseDatasetParamsFromPost(c *gin.Context) (*ReqDatasetParams, error) {

	var params ReqDatasetParams

	err := c.Bind(&params)

	if err != nil {
		return nil, err
	}

	return &params, nil
}

// If species is the empty string, species will be determined
// from the url parameters
// func GeneInfoRoute(c *gin.Context, species string) error {
// 	if species == "" {
// 		species = c.Param("species")
// 	}

// 	params, err := ParseParamsFromPost(c)

// 	if err != nil {
// 		return web.ErrorReq(err)
// 	}

// 	ret := make([]geneconv.Conversion, len(params.Searches))

// 	for ni, search := range params.Searches {

// 		genes, _ := geneconvdb.GeneInfo(search, species, params.Exact)

// 		ret[ni] = geneconv.Conversion{Search: search, Genes: genes}
// 	}

// 	web.MakeDataResp(c, "", ret)
// }

func GenesRoute(c *gin.Context) {
	genes, err := pathwaydb.GeneList()

	if err != nil {
		c.Error(err)
		return
	}

	web.MakeDataResp(c, "", genes)
}

func DatasetsInfoRoute(c *gin.Context) {

	datasets, err := pathwaydb.AllDatasetsInfo()

	if err != nil {
		c.Error(err)
		return
	}

	web.MakeDataResp(c, "", datasets)
}

// Returns the pathways for a specific collection.
// The collection id is expected to be in the url parameter "id",
// e.g. /pathway/collections/1234 would return the pathways for collection with id 1234.
// ids are uuidv7s, so they are 36 characters long and contain hyphens.
// If the id parameter is missing or empty, a bad request response is returned.
func CollectionRoute(c *gin.Context) {

	id := c.Param("id")

	if id == "" {
		web.BadReqResp(c, errors.New("id parameter is required"))
		return
	}

	collection, err := pathwaydb.GetCollection(id)

	if err != nil {
		c.Error(err)
		return
	}

	web.MakeDataResp(c, "", collection)
}

// Returns the set of pathways belonging to the specified collections.
// The collection ids are expected to be in the post body as a json array of
// strings with the key "ids". ids are uuidv7s, so they are 36 characters
// long and contain hyphens. Data is returned as a list of datasets,
// each with a list of collections, each with a list of pathways.
func CollectionsRoute(c *gin.Context) {

	params, err := web.ParseIdParamsFromPost(c)

	if err != nil {
		c.Error(err)
		return
	}

	datasets, err := pathwaydb.GetCollections(params.Ids)

	if err != nil {
		c.Error(err)
		return
	}

	web.MakeDataResp(c, "", datasets)
}

func PathwaysRoute(c *gin.Context) {

	params, err := web.ParseIdParamsFromPost(c)

	if err != nil {
		c.Error(err)
		return
	}

	datasets, err := pathwaydb.GetGeneSets(params.Ids)

	if err != nil {
		c.Error(err)
		return
	}

	web.MakeDataResp(c, "", datasets)
}

func PathwayOverlapRoute(c *gin.Context) {

	params, err := ParseOverlapParamsFromPost(c)

	if err != nil {
		c.Error(err)
		return
	}

	//testPathway := params.Geneset.ToPathway()

	tests, err := pathwaydb.Overlap(params.Geneset, params.Datasets)

	if err != nil {
		c.Error(err)
		return
	}

	//var ret geneconv.ConversionResults

	//ret.Conversions = make([]geneconv.Conversion, len(params.Searches))

	// for _, search := range params.Genes {

	// 	// Don't care about the errors, just plug empty list into failures
	// 	//conversion, _ := geneconvdbcache.Convert(search, fromSpecies, toSpecies, params.Exact)

	// 	ret.Conversions = append(ret.Conversions, conversion)
	// }

	web.MakeDataResp(c, "", tests)

	// web.MakeDataResp(c, "", mutationdbcache.GetInstance().List())
}
