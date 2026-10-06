package database

import (
	"context"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pardnchiu/PodRun/internal/model"
)

func (s *SQLite) UpdatePod(ctx context.Context, d *model.Pod) error {
	_, err := s.db.ExecContext(ctx, `
  UPDATE pods
  SET
    status = ?,
    updated_at = CURRENT_TIMESTAMP,
    dismiss = ?
  WHERE uid = ?
  `,
		d.Status,
		d.Dismiss,
		d.UID,
	)
	return err
}
