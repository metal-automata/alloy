package store

import (
	"context"

	"github.com/metal-automata/alloy/internal/app"
	"github.com/metal-automata/alloy/internal/model"
	"github.com/metal-automata/alloy/internal/store/csv"
	"github.com/metal-automata/alloy/internal/store/fleetdb"
	"github.com/metal-automata/alloy/internal/store/mock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	ErrStore = errors.New("store error")
)

type Repository interface {
	// Kind returns the repository store kind.
	Kind() model.StoreKind

	// AssetByID returns one asset from the inventory identified by its identifier.
	AssetByID(ctx context.Context, assetID string, fetchBmcCredentials bool) (*model.Asset, error)

	AssetsByOffsetLimit(ctx context.Context, offset, limit int) (assets []*model.Asset, totalAssets int, err error)

	// AssetUpdate inserts and updates collected data for the asset in the store.
	AssetUpdate(ctx context.Context, asset *model.Asset) error
}

func NewRepository(ctx context.Context, storeKind model.StoreKind, appKind model.AppKind, cfg *app.Configuration, logger *logrus.Logger) (Repository, error) {
	switch storeKind {
	case model.StoreKindFleetDB:
		return fleetdb.New(ctx, appKind, cfg.FleetDBAPIOptions, logger)

	case model.StoreKindCsv:
		return csv.New(ctx, cfg.CsvFile, logger)

	case model.StoreKindMock:
		assets := 10
		return mock.New(assets)

	default:
		return nil, errors.Wrap(ErrStore, "unsupported store kind: "+string(storeKind))
	}
}
