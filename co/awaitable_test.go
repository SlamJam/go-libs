package co_test

import (
	"context"

	"github.com/SlamJam/go-libs/co"
	"github.com/pkg/errors"
)

type PropertyID int
type Property struct {
	OtelloID int
	RegionID RegionID
}
type Room struct {
	Title string
}

type RegionID int
type Region struct {
}

func GetPropByID(context.Context, PropertyID) (Property, error) {
	return Property{}, nil
}

func GetRooms(context.Context, PropertyID) ([]Room, error) {
	return []Room{}, nil
}

func GetRegion(context.Context, RegionID) (Region, error) {
	return Region{}, nil
}

func _() {
	// во внешнем пакете такое будет сделать нельзя
	// All(&co.promise[int]{})

	co.All(context.TODO(), co.Resolved)

	ctx := context.Background()

	var propID PropertyID = 1

	propetryP := co.NewPromise(func() (Property, error) {
		return GetPropByID(ctx, propID)
	})
	roomsP := co.NewPromise(func() ([]Room, error) {
		return GetRooms(ctx, propID)
	})

	regionP := co.NewPromise(func() (Region, error) {
		prop, err := propetryP.Poll(ctx)
		if err != nil {
			return Region{}, errors.Wrap(err, "fail load region")
		}

		return GetRegion(ctx, prop.RegionID)
	})

	if err := co.All(context.TODO(), propetryP, roomsP, regionP); err != nil {
		return
	}

	prop := propetryP.Value()
	_ = prop.OtelloID

	rooms := roomsP.Value()
	_ = len(rooms)

	region := regionP.Value()
	_ = region
}
