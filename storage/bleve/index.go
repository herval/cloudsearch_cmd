package bleve

import (
	"github.com/herval/cloudsearch"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/char/html"
	"github.com/blevesearch/bleve/analysis/datetime/optional"
	"github.com/blevesearch/bleve/analysis/token/camelcase"
	"github.com/blevesearch/bleve/analysis/token/lowercase"
	"github.com/blevesearch/bleve/analysis/tokenizer/web"
	"os"
)

func NewIndex(storagePath string, version string) (bleve.Index, error) {
	var err error
	path := cloudsearch.FileAt(storagePath, fmt.Sprintf("index" + version + ".bleve"))
	mapping := bleve.NewIndexMapping()

	lowerCase := bleve.NewTextFieldMapping()
	lowerCase.Analyzer = simple.Name

	//tokenizers
	resultAnalyser := "resultAnalyser"
	//edgengramNFC := "edgeNgram325"
	//err = mapping.AddCustomTokenFilter(edgengramNFC,
	//	map[string]interface{}{
	//		"type": edgengram.Name,
	//		"min":  3.0,
	//		"max":  25.0,
	//	})
	//if err != nil {
	//	return nil, err
	//}
	if err = mapping.AddCustomAnalyzer(resultAnalyser, map[string]interface{}{
		"type":          custom.Name,
		"char_filters":  []string{html.Name},
		"tokenizer":     web.Name,
		"token_filters": []string{camelcase.Name, lowercase.Name},
	}); err != nil {
		return nil, err
	}

	// field mapping types
	keywordContent := bleve.NewTextFieldMapping()
	keywordContent.Analyzer = resultAnalyser

	simpleContent := bleve.NewTextFieldMapping()
	simpleContent.Analyzer = simple.Name

	dateTime := bleve.NewDateTimeFieldMapping()

	// bundle the entire thing together
	d := bleve.NewDocumentMapping()
	d.AddFieldMappingsAt("ContentType", lowerCase)
	d.AddFieldMappingsAt("Title", keywordContent)
	d.AddFieldMappingsAt("Permalink", lowerCase, simpleContent)
	d.AddFieldMappingsAt("Body", keywordContent)
	d.AddFieldMappingsAt("Timestamp", dateTime)

	mapping.AddDocumentMapping("searchableResult", d)
	mapping.DefaultDateTimeParser = optional.Name
	//mapping.DefaultAnalyzer = resultAnalyser

	// this is defaulted to "searchableResult" for now, we can have more types maybe?
	mapping.TypeField = "Type"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return bleve.New(path, mapping)
	} else {
		return bleve.Open(path)
	}
}
