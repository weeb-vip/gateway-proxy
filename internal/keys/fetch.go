package keys

import (
	"errors"
	"github.com/machinebox/graphql"
)

const FetchKeysDocument = `
query FetchAllKeys{
  keys{
    id
    body
  }
}
`

func (r keyFetcher) FetchKeys() ([]Key, error) {
	response, err := run[GraphQLResponse](graphql.NewClient(r.graphqlEndpoint), graphql.NewRequest(FetchKeysDocument))
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, errors.New("nil response returned")
	}

	return response.Keys, nil
}

func NewFetcher(endpoint string) Fetcher {
	return keyFetcher{graphqlEndpoint: endpoint}
}
