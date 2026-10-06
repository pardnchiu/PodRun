package database

import (
	"context"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pardnchiu/PodRun/internal/model"
)

func (s *SQLite) UpsertPod(ctx context.Context, d *model.Pod) error {
	_, err := s.db.ExecContext(ctx, `
  INSERT INTO pods (
    uid, pod_uid, pod_name, local_dir, remote_dir,
    file, target, status, hostname, ip,
    replicas
  )
  VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?
  )
  ON CONFLICT(uid) DO UPDATE SET
    pod_name = excluded.pod_name,
    local_dir = excluded.local_dir,
    remote_dir = excluded.remote_dir,
    file = excluded.file,
    target = excluded.target,
    status = excluded.status,
    hostname = excluded.hostname,
    ip = excluded.ip,
    replicas = excluded.replicas,
    updated_at = CURRENT_TIMESTAMP,
    dismiss = 0
  `,
		d.UID,
		d.PodID,
		d.PodName,
		d.LocalDir,
		d.RemoteDir,
		d.File,
		d.Target,
		d.Status,
		d.Hostname,
		d.IP,
		d.Replicas,
	)
	return err
}
