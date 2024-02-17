package app

import (
	"context"
	"fmt"

	"github.com/takaidohigasi/skeefree/go/config"
	"github.com/takaidohigasi/skeefree/go/core"
	"github.com/takaidohigasi/skeefree/go/db"
	"github.com/takaidohigasi/skeefree/go/gh"
	"github.com/takaidohigasi/skeefree/go/util"

	"github.com/uber-go/zap"
	"github.com/github/mu/logger"
)

// DirectApplier applies "direct" migrations in "ready" state
type DirectApplier struct {
	cfg               *config.Config
	logger            *logger.Logger
	backend           *db.Backend
	mysqlDiscoveryAPI *gh.MySQLDiscoveryAPI
}

// NewDirectApplier creates a new direct applier object
func NewDirectApplier(c *config.Config, logger *logger.Logger, backend *db.Backend, mysqlDiscoveryAPI *gh.MySQLDiscoveryAPI) *DirectApplier {
	return &DirectApplier{
		cfg:               c,
		logger:            logger,
		backend:           backend,
		mysqlDiscoveryAPI: mysqlDiscoveryAPI,
	}
}

func (applier *DirectApplier) applyNextMigration(ctx context.Context, onOwned, onRunning, onComplete, onFailed func(m *core.Migration)) (migration *core.Migration, err error) {
	applier.logger.Log(ctx, "direct-applier: applyNextMigration", zap.String("service-id", applier.backend.ServiceId()))
	token := util.PrettyUniqueToken()
	if migration, err = applier.backend.OwnReadyMigration(core.MigrationStrategyDirect, token); err != nil {
		return migration, err
	}
	if migration == nil {
		applier.logger.Log(ctx, "direct-applier: no migration owned")
		return migration, nil
	}
	onOwned(migration)
	applier.logger.Log(ctx, "direct-applier: migration owned", zap.Any("pr", migration.PR), zap.String("canonical", migration.Canonical), zap.String("strategy", string(migration.Strategy)))
	if _, err := applier.backend.UpdateMigrationStatus(migration, core.MigrationStatusReady, core.MigrationStatusRunning, core.MigrationStrategyDirect); err != nil {
		return migration, err
	}
	migration.Cluster, err = applier.mysqlDiscoveryAPI.GetCluster(migration.Cluster.Name)
	if err != nil {
		return migration, err
	}
	applier.logger.Log(ctx, "direct-applier: migration cluster", zap.Any("pr", migration.PR), zap.String("canonical", migration.Canonical), zap.String("cluster", migration.Cluster.Name), zap.String("rw", migration.Cluster.RWName))

	topology, err := db.NewTopologyDB(applier.cfg, migration)
	if err != nil {
		return migration, err
	}
	onRunning(migration)
	// friendly health check
	readOnly, err := topology.Ping()
	if err != nil {
		return migration, err
	}
	if readOnly {
		return migration, fmt.Errorf("Attempt to run migration: host found to be read only for `%s/%s` via `%s`", migration.Cluster.Name, migration.Repo.MySQLSchema, migration.Cluster.RWName)
	}
	applier.logger.Log(ctx, "direct-applier: ping", zap.Any("pr", migration.PR), zap.String("canonical", migration.Canonical), zap.Any("read_only", readOnly))
	// Actually run the statement:
	if _, err := topology.Exec(migration.PRStatement.Statement); err != nil {
		applier.backend.UpdateMigrationStatus(migration, core.MigrationStatusRunning, core.MigrationStatusFailed, core.MigrationStrategyDirect)
		onFailed(migration)
		return migration, err
	}

	applier.backend.UpdateMigrationStatus(migration, core.MigrationStatusRunning, core.MigrationStatusComplete, core.MigrationStrategyDirect)
	onComplete(migration)
	return migration, nil
}
