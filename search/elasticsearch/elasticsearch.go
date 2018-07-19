package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/olivere/elastic"
)

type Config struct {
	Host  string
	Index string
}

type SearchClient struct {
	config  Config
	eclient *elastic.Client
}

// New Creates a new SearchClient and initializes and elasticsearch client
// in the process.
func New(conf Config) (*SearchClient, error) {
	// Set sniff seems to be required when running elastic search in single node mode
	client, err := elastic.NewClient(elastic.SetURL(conf.Host), elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}

	// Attempt to ping the es client
	esversion, err := client.ElasticsearchVersion(conf.Host)
	if err != nil {
		return nil, err
	}
	log.Printf("Succesfully connected to elasticsearch version %v", esversion)

	sclient := &SearchClient{
		config:  conf,
		eclient: client,
	}

	if err := sclient.InitializeIndex(conf.Index); err != nil {
		return nil, err
	}

	return sclient, nil
}

// Search consumes a query string and finds matching documents in ElasticSearch.
func (s *SearchClient) Search(query string, orgs []string) ([]question.Question, error) {
	log.Printf("received search for query: %v in orgs: %v", query, orgs)

	ctx := context.Background()

	boolQuery := elastic.NewBoolQuery()

	boolQuery.Must(elastic.NewMultiMatchQuery(query, "title", "content"))

	if len(orgs) > 0 {
		terms := make([]interface{}, len(orgs))
		for i, org := range orgs {
			// Term query performs a non-analyzed search. By default all tokens
			// are stored in lowercase.
			terms[i] = strings.ToLower(org)
		}

		boolQuery.Must(elastic.NewTermsQuery("organization", terms...))
	}

	searchResult, err := s.eclient.Search().
		Index(s.config.Index).
		Query(boolQuery).Do(ctx)
	if err != nil {
		return nil, err
	}

	questions := make([]question.Question, 0)
	var qt question.Question
	for _, item := range searchResult.Each(reflect.TypeOf(qt)) {
		if q, ok := item.(question.Question); ok {
			questions = append(questions, q)
		}
	}

	return questions, nil
}

// IndexQuestion consumes a quesiton and inserts it into elasticsearch.
func (s *SearchClient) IndexQuestion(q question.Question) error {
	ctx := context.Background()

	// Inserts the given question into elastic search to make it searchable
	_, err := s.eclient.Index().
		Index(s.config.Index).
		Type("question").
		Id(fmt.Sprintf("%v", q.ID)).
		BodyJson(q).
		Do(ctx)
	if err != nil {
		return err
	}

	// Flush to make sure the documents got written.
	// TODO: This should be ran periodically in the background instead -
	// i.e every 30s call Flush()
	_, err = s.eclient.Flush().Index(s.config.Index).Do(ctx)
	if err != nil {
		return err
	}

	log.Printf("Inserted question to elasticsearch")

	return nil
}

// InitializeIndex ensures the procided index name exists in ES
// otherwise it creates it.
func (s *SearchClient) InitializeIndex(index string) error {
	ctx := context.Background()

	// Use the IndexExists service to check if a specified index exists.
	exists, err := s.eclient.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		// Create a new index.
		createIndex, err := s.eclient.CreateIndex(index).Do(ctx)
		if err != nil {
			return err
		}

		if !createIndex.Acknowledged {
			return errors.New("unable to initialize index")
		}
	}

	return nil
}
