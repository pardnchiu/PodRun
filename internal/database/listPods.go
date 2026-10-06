package database

import (
	"context"

	"github.com/pardnchiu/PodRun/internal/model"
)

func (s *SQLite) ListPods(ctx context.Context) ([]model.Pod, error) {
	rows, err := s.db.QueryContext(ctx, `
	SELECT
	  id, uid, pod_uid, pod_name, local_dir,
		remote_dir, file, target, status, hostname,
		ip, replicas, created_at, updated_at
	FROM pods
	WHERE dismiss = 0
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []model.Pod
	for rows.Next() {
		var c model.Pod
		if err := rows.Scan(
			&c.ID, &c.UID, &c.PodID, &c.PodName, &c.LocalDir,
			&c.RemoteDir, &c.File, &c.Target, &c.Status, &c.Hostname,
			&c.IP, &c.Replicas, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		containers = append(containers, c)
	}

	return containers, rows.Err()
}
