package co_test

import (
	"context"
	"testing"

	"github.com/SlamJam/go-libs/co"
)

type ProviderID int
type Offer struct {
}

func GetOffersProviderA(context.Context) ([]Offer, error) {
	return []Offer{}, nil
}

func GetOffersProviderB(context.Context) ([]Offer, error) {
	return []Offer{}, nil
}

func TestXxx(t *testing.T) {
	var pm co.PromiseMap[ProviderID, []Offer]

	var (
		ProviderA ProviderID = 5
		ProviderB ProviderID = 6
	)

	ctx := context.Background()
	pctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pm.Add(ProviderA, func() ([]Offer, error) {
		return GetOffersProviderA(pctx)
	})

	pm.Add(ProviderB, func() ([]Offer, error) {
		return GetOffersProviderB(pctx)
	})

	res, err := pm.AllResults(ctx)
	if err != nil {
		t.Fatal()
	}

	t.Log("FOO", res)
}
