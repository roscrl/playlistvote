package domain

type PlaylistIDUpvoteMessage struct {
	PlaylistID PlaylistID
	Upvotes    int64
	UpvotedBy  SessionID
}
