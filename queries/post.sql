-- name: GetPublicPosts :many
SELECT * from post;

-- name: CreatePost :one
INSERT INTO post (id, title, body)
VALUES (uuid_generate_v4(),$1, $2)
RETURNING id, title, body;

-- name: GetPostByID :one
SELECT id, title, body FROM post
WHERE id = $1;