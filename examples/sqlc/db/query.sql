-- name: ListUsers :many
SELECT *
FROM users
ORDER BY id
LIMIT ?1 OFFSET ?2;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES (?, ?) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET name = ?, email = ? WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: SearchUsers :many
-- name: SearchUsers :many
SELECT
  id,
  name,
  email,
  created_at
FROM users
WHERE
  (sqlc.narg(id) IS NULL OR id = sqlc.narg(id))
  AND (
    sqlc.narg(name) IS NULL
    OR lower(name) LIKE '%' || lower(sqlc.narg(name)) || '%'
  )
  AND (
    sqlc.narg(email) IS NULL
    OR lower(email) LIKE '%' || lower(sqlc.narg(email)) || '%'
  )
  AND (
    sqlc.narg(created_at) IS NULL
    OR created_at = sqlc.narg(created_at)
  )
ORDER BY created_at DESC;