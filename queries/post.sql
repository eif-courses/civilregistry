-- name: GetPublicPosts :many
SELECT * from post;

-- name: CreatePost :one
INSERT INTO post (title, body)
VALUES ($1, $2)
RETURNING id, title, body;

-- name: GetPostByID :one
SELECT id, title, body FROM post
WHERE id = $1;