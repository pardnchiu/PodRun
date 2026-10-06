package database

import (
	"context"

	"github.com/pardnchiu/PodRun/internal/model"
)

func (s *SQLite) PodInfo(ctx context.Context, uid string) (*model.Pod, error) {
	row := s.db.QueryRowContext(ctx, `
  SELECT
    id, uid, pod_uid, pod_name, local_dir,
    remote_dir, file, target, status, hostname,
    ip, replicas, created_at, updated_at
  FROM pods
  WHERE dismiss = 0 AND uid = ?
  LIMIT 1
  `, uid)

	var c model.Pod
	err := row.Scan(
		&c.ID, &c.UID, &c.PodID, &c.PodName, &c.LocalDir,
		&c.RemoteDir, &c.File, &c.Target, &c.Status, &c.Hostname,
		&c.IP, &c.Replicas, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
