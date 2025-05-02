package keys

import (
	"context"
	"github.com/machinebox/graphql"
)

func run[T any](client *graphql.Client, request *graphql.Request) (*T, error) {
	response := new(T)

	err := client.Run(context.Background(), request, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
