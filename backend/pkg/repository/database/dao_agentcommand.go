package database

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *daoRepo) ListAgentCommands(ctx core.Context, node, probe string, limit int) ([]ac.Command, error) {
	var commands []ac.Command
	err := repo.GetContextDB(ctx).Where("node_id=? AND probe_type=?", node, probe).Order("created_at desc").Limit(limit).Find(&commands).Error
	return commands, err
}

func (repo *daoRepo) ClaimAgentCommands(ctx core.Context, node, probe, claimant string, now time.Time, ttl time.Duration) ([]ac.Command, error) {
	var out []ac.Command
	err := repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		var pending []ac.Command
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("node_id=? AND probe_type=? AND (status=? OR (status=? AND lease_expires_at < ?))", node, probe, "PENDING", "CLAIMED", now).Order("created_at asc").Limit(20).Find(&pending).Error; e != nil {
			return e
		}
		for i := range pending {
			token := uuid.NewString()
			expires := now.Add(ttl)
			res := tx.Model(&ac.Command{}).Where("id=? AND (status=? OR (status=? AND lease_expires_at < ?))", pending[i].ID, "PENDING", "CLAIMED", now).Updates(map[string]any{"status": "CLAIMED", "claimant_id": claimant, "lease_token": token, "lease_expires_at": expires, "delivery_count": gorm.Expr("delivery_count + 1")})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				continue
			}
			if e := tx.First(&pending[i], "id=?", pending[i].ID).Error; e != nil {
				return e
			}
			out = append(out, pending[i])
		}
		return nil
	})
	return out, err
}

func (repo *daoRepo) TransitionAgentCommand(ctx core.Context, id uuid.UUID, node, claimant, token, status, result string, now time.Time) (*ac.Command, error) {
	var command ac.Command
	err := repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&command, "id=? AND node_id=?", id, node).Error; e != nil {
			return e
		}
		if ac.IsCommandTerminal(command.Status) {
			if command.Status == status {
				return nil
			}
			return fmt.Errorf("command already terminal: %s", command.Status)
		}
		if command.ClaimantID != claimant || command.LeaseToken != token {
			return fmt.Errorf("command lease mismatch")
		}
		updates := map[string]any{"status": status, "result": result, "updated_at": now}
		if err := ac.ValidateCommandTransition(command.Status, status); err != nil {
			return err
		}
		switch status {
		case "STARTED":
			if command.StartedAt == nil {
				updates["started_at"] = now
			}
		case "SUCCEEDED", "FAILED":
			updates["finished_at"] = now
		}
		if e := tx.Model(&command).Updates(updates).Error; e != nil {
			return e
		}
		return tx.First(&command, "id=?", id).Error
	})
	return &command, err
}
