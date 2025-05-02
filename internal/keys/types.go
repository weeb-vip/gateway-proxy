package keys

type Fetcher interface {
	FetchKeys() ([]Key, error)
}

type keyFetcher struct {
	graphqlEndpoint string
}

type GraphQLResponse struct {
	Keys []Key `json:"keys"`
}

type Key struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}
