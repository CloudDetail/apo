package database

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AgentConfigRepo interface {
	CreateAgentTemplate(core.Context, *ac.Template) error
	UpdateAgentTemplate(core.Context, *ac.Template) error
	DeleteAgentTemplate(core.Context, uuid.UUID) error
	GetAgentTemplate(core.Context, uuid.UUID) (*ac.Template, error)
	ListAgentTemplates(core.Context) ([]ac.Template, error)
	UpsertAgentNode(core.Context, *ac.Node) error
	ListAgentNodes(core.Context) ([]ac.Node, error)
	GetAgentNode(core.Context, string) (*ac.Node, error)
	SaveAgentInitialConfigIfUnassigned(core.Context, *ac.InitialConfig) (bool, error)
	GetAgentInitialConfig(core.Context, string, string) (*ac.InitialConfig, error)
	ListAgentInitialConfigs(core.Context, string) ([]ac.InitialConfig, error)
	GetGlobalCircuitConfig(core.Context) (*ac.GlobalCircuitConfig, error)
	SaveGlobalCircuitConfig(core.Context, *ac.GlobalCircuitConfig) error
	GetNodeCircuitConfig(core.Context, string) (*ac.NodeCircuitConfig, error)
	SaveNodeCircuitConfigs(core.Context, []*ac.NodeCircuitConfig) error
	GetNodeLogConfig(core.Context, string, string) (*ac.NodeLogConfig, error)
	SaveNodeLogConfigs(core.Context, []*ac.NodeLogConfig) error
	GetLogCollectionConfig(core.Context, string, string) (*ac.LogCollectionConfig, error)
	SaveLogCollectionConfig(core.Context, *ac.LogCollectionConfig) error
	SaveAgentBinding(core.Context, *ac.Binding, *ac.Release) error
	SaveAgentBindings(core.Context, []*ac.Binding, []*ac.Release, map[string]int64) error
	GetAgentRelease(core.Context, string, string) (*ac.Release, error)
	ListAgentReleases(core.Context, string, string, int) ([]ac.Release, error)
	ListLatestAgentReleases(core.Context, string) ([]ac.Release, error)
	ReportAgentState(core.Context, *ac.ApplyState) error
	ListAgentStates(core.Context) ([]ac.ApplyState, error)
	CreateAgentCommand(core.Context, *ac.Command) error
	ListAgentCommands(core.Context, string, string, int) ([]ac.Command, error)
	ClaimAgentCommands(core.Context, string, string, string, time.Time, time.Duration) ([]ac.Command, error)
	TransitionAgentCommand(core.Context, uuid.UUID, string, string, string, string, string, time.Time) (*ac.Command, error)
}

func (repo *daoRepo) CreateAgentTemplate(ctx core.Context, t *ac.Template) error {
	return repo.GetContextDB(ctx).Create(t).Error
}
func (repo *daoRepo) UpdateAgentTemplate(ctx core.Context, t *ac.Template) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Model(&ac.Template{}).Where("id = ?", t.ID).Omit("Variables").Updates(t).Error; e != nil {
			return e
		}
		if e := tx.Where("template_id = ?", t.ID).Delete(&ac.TemplateVariable{}).Error; e != nil {
			return e
		}
		if len(t.Variables) > 0 {
			return tx.Create(&t.Variables).Error
		}
		return nil
	})
}
func (repo *daoRepo) DeleteAgentTemplate(ctx core.Context, id uuid.UUID) error {
	return repo.GetContextDB(ctx).Delete(&ac.Template{}, "id = ?", id).Error
}
func (repo *daoRepo) GetAgentTemplate(ctx core.Context, id uuid.UUID) (*ac.Template, error) {
	var t ac.Template
	e := repo.GetContextDB(ctx).Preload("Variables").First(&t, "id = ?", id).Error
	return &t, e
}
func (repo *daoRepo) ListAgentTemplates(ctx core.Context) ([]ac.Template, error) {
	var ts []ac.Template
	q := repo.GetContextDB(ctx).Preload("Variables")
	e := q.Order("updated_at desc").Find(&ts).Error
	return ts, e
}
func (repo *daoRepo) UpsertAgentNode(ctx core.Context, n *ac.Node) error {
	return repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "cluster_id", "environment", "labels", "daemon_version",
			"last_seen_at", "updated_at",
		}),
	}).Create(n).Error
}
func (repo *daoRepo) ListAgentNodes(ctx core.Context) ([]ac.Node, error) {
	var ns []ac.Node
	e := repo.GetContextDB(ctx).Order("updated_at desc").Find(&ns).Error
	return ns, e
}
func (repo *daoRepo) GetAgentNode(ctx core.Context, id string) (*ac.Node, error) {
	var n ac.Node
	e := repo.GetContextDB(ctx).First(&n, "id = ?", id).Error
	return &n, e
}
func (repo *daoRepo) SaveAgentInitialConfigIfUnassigned(ctx core.Context, baseline *ac.InitialConfig) (bool, error) {
	saved := false
	err := repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&ac.Release{}).Where("node_id=? AND probe_type=?", baseline.NodeID, baseline.ProbeType).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		result := tx.Where("node_id=? AND probe_type=?", baseline.NodeID, baseline.ProbeType).FirstOrCreate(baseline)
		if result.Error != nil {
			return result.Error
		}
		saved = result.RowsAffected > 0
		return nil
	})
	return saved, err
}
func (repo *daoRepo) GetAgentInitialConfig(ctx core.Context, node, probe string) (*ac.InitialConfig, error) {
	var baseline ac.InitialConfig
	err := repo.GetContextDB(ctx).Where("node_id=? AND probe_type=?", node, probe).First(&baseline).Error
	return &baseline, err
}
func (repo *daoRepo) ListAgentInitialConfigs(ctx core.Context, node string) ([]ac.InitialConfig, error) {
	var baselines []ac.InitialConfig
	err := repo.GetContextDB(ctx).Where("node_id=?", node).Order("probe_type").Find(&baselines).Error
	return baselines, err
}
func (repo *daoRepo) GetGlobalCircuitConfig(ctx core.Context) (*ac.GlobalCircuitConfig, error) {
	var config ac.GlobalCircuitConfig
	err := repo.GetContextDB(ctx).First(&config, 1).Error
	return &config, err
}
func (repo *daoRepo) SaveGlobalCircuitConfig(ctx core.Context, config *ac.GlobalCircuitConfig) error {
	config.ID = 1
	return repo.GetContextDB(ctx).Save(config).Error
}
func (repo *daoRepo) GetNodeCircuitConfig(ctx core.Context, node string) (*ac.NodeCircuitConfig, error) {
	var config ac.NodeCircuitConfig
	err := repo.GetContextDB(ctx).Where("node_id=?", node).Order("updated_at desc").First(&config).Error
	return &config, err
}
func (repo *daoRepo) SaveNodeCircuitConfigs(ctx core.Context, configs []*ac.NodeCircuitConfig) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			var old ac.NodeCircuitConfig
			err := tx.Where("node_id=?", config.NodeID).Order("updated_at desc").First(&old).Error
			if err == nil {
				config.ID = old.ID
				config.CreatedAt = old.CreatedAt
				config.Version = old.Version + 1
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Save(config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (repo *daoRepo) GetNodeLogConfig(ctx core.Context, node, probe string) (*ac.NodeLogConfig, error) {
	var config ac.NodeLogConfig
	err := repo.GetContextDB(ctx).Where("node_id=? AND probe_type=?", node, probe).First(&config).Error
	return &config, err
}
func (repo *daoRepo) SaveNodeLogConfigs(ctx core.Context, configs []*ac.NodeLogConfig) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			var old ac.NodeLogConfig
			err := tx.Where("node_id=? AND probe_type=?", config.NodeID, config.ProbeType).First(&old).Error
			if err == nil {
				config.ID = old.ID
				config.CreatedAt = old.CreatedAt
				config.Version = old.Version + 1
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Save(config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (repo *daoRepo) GetLogCollectionConfig(ctx core.Context, scopeType, scopeID string) (*ac.LogCollectionConfig, error) {
	var config ac.LogCollectionConfig
	err := repo.GetContextDB(ctx).Where("scope_type=? AND scope_id=?", scopeType, scopeID).First(&config).Error
	return &config, err
}
func (repo *daoRepo) SaveLogCollectionConfig(ctx core.Context, config *ac.LogCollectionConfig) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		var old ac.LogCollectionConfig
		err := tx.Where("scope_type=? AND scope_id=?", config.ScopeType, config.ScopeID).First(&old).Error
		if err == nil {
			config.ID = old.ID
			config.CreatedAt = old.CreatedAt
			config.Version = old.Version + 1
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return tx.Save(config).Error
	})
}
func (repo *daoRepo) SaveAgentBinding(ctx core.Context, b *ac.Binding, r *ac.Release) error {
	return repo.SaveAgentBindings(ctx, []*ac.Binding{b}, []*ac.Release{r}, map[string]int64{agentReleaseKey(r.NodeID, r.ProbeType): r.Generation - 1})
}
func agentReleaseKey(node, probe string) string { return node + "\x00" + ac.NormalizeProbe(probe) }
func (repo *daoRepo) SaveAgentBindings(ctx core.Context, bindings []*ac.Binding, releases []*ac.Release, expected map[string]int64) error {
	if len(bindings) != len(releases) {
		return errors.New("binding and release counts do not match")
	}
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		keys := make([]string, 0, len(expected))
		nodeSet := make(map[string]struct{})
		for key := range expected {
			parts := strings.SplitN(key, "\x00", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid expected generation key")
			}
			keys = append(keys, key)
			nodeSet[parts[0]] = struct{}{}
		}
		nodeIDs := make([]string, 0, len(nodeSet))
		for nodeID := range nodeSet {
			nodeIDs = append(nodeIDs, nodeID)
		}
		sort.Strings(nodeIDs)
		for _, nodeID := range nodeIDs {
			var node ac.Node
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", nodeID).First(&node).Error; err != nil {
				return err
			}
		}
		sort.Strings(keys)
		for _, key := range keys {
			parts := strings.SplitN(key, "\x00", 2)
			nodeID, probeType := parts[0], parts[1]
			var current ac.Release
			err := tx.Where("node_id=? AND probe_type=?", nodeID, probeType).Order("generation desc").First(&current).Error
			actual := int64(0)
			if err == nil {
				actual = current.Generation
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			want := expected[key]
			if actual != want {
				return fmt.Errorf("generation conflict for %s/%s: expected %d, actual %d", nodeID, probeType, want, actual)
			}
		}
		for index, binding := range bindings {
			var old ac.Binding
			err := tx.Where("node_id=? AND probe_type=?", binding.NodeID, binding.ProbeType).First(&old).Error
			if err == nil {
				binding.ID = old.ID
				binding.CreatedAt = old.CreatedAt
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Save(binding).Error; err != nil {
				return err
			}
			if err := tx.Create(releases[index]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (repo *daoRepo) GetAgentRelease(ctx core.Context, node, probe string) (*ac.Release, error) {
	var r ac.Release
	e := repo.GetContextDB(ctx).Where("node_id=? AND probe_type=?", node, probe).Order("generation desc").First(&r).Error
	return &r, e
}
func (repo *daoRepo) ListAgentReleases(ctx core.Context, node, probe string, limit int) ([]ac.Release, error) {
	var releases []ac.Release
	err := repo.GetContextDB(ctx).Where("node_id=? AND probe_type=?", node, probe).Order("generation desc").Limit(limit).Find(&releases).Error
	return releases, err
}
func (repo *daoRepo) ListLatestAgentReleases(ctx core.Context, node string) ([]ac.Release, error) {
	var releases []ac.Release
	if err := repo.GetContextDB(ctx).Where("node_id=?", node).Order("probe_type, generation desc").Find(&releases).Error; err != nil {
		return nil, err
	}
	latest := make([]ac.Release, 0, len(releases))
	seen := map[string]bool{}
	for _, release := range releases {
		if !seen[release.ProbeType] {
			seen[release.ProbeType] = true
			latest = append(latest, release)
		}
	}
	return latest, nil
}
func (repo *daoRepo) ReportAgentState(ctx core.Context, s *ac.ApplyState) error {
	now := time.Now()
	s.ReportedAt = &now
	s.UpdatedAt = now
	return repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "node_id"}, {Name: "probe_type"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"desired_generation", "applied_generation", "status", "reason", "error",
			"runtime", "circuit", "reported_at", "updated_at",
		}),
	}).Create(s).Error
}
func (repo *daoRepo) ListAgentStates(ctx core.Context) ([]ac.ApplyState, error) {
	var states []ac.ApplyState
	err := repo.GetContextDB(ctx).Order("node_id, probe_type").Find(&states).Error
	return states, err
}
func (repo *daoRepo) CreateAgentCommand(ctx core.Context, c *ac.Command) error {
	return repo.GetContextDB(ctx).Create(c).Error
}
