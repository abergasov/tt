package fetch

import "context"

//go:generate mockgen -source=abstract.go -destination=abstract_fetch_mock.go -package=fetch
type Fetcher interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}
