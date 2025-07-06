-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
	INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
	)
	RETURNING *
)
SELECT
	inserted_feed_follow.*,
	feeds.name AS feed_name,
	users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN users ON users.id = inserted_feed_follow.user_id;

-- name: GetFeedFollowsForUser :many
WITH follows_for_user AS (
	SELECT * FROM feed_follows WHERE feed_follows.user_id = $1
)
SELECT follows_for_user.*, feeds.name AS feed_name, users.name AS user_name FROM follows_for_user
INNER JOIN feeds ON follows_for_user.feed_id = feeds.id 
INNER JOIN users ON follows_for_user.user_id = users.id;
