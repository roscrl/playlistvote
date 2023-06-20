package broadcast

import (
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"app/core/domain"
	"app/core/session"
)

type UpvoteBroadcasterService struct {
	playlistsBeingUpvoted []*domain.PlaylistIDUpvoteMessage
	broadcastQueue        chan *domain.PlaylistIDUpvoteMessage
	lock                  sync.Mutex
}

func NewUpvoteBroadcasterService() *UpvoteBroadcasterService {
	const broadcastQueueSize = 1000

	return &UpvoteBroadcasterService{
		playlistsBeingUpvoted: []*domain.PlaylistIDUpvoteMessage{},
		broadcastQueue:        make(chan *domain.PlaylistIDUpvoteMessage, broadcastQueueSize),
		lock:                  sync.Mutex{},
	}
}

func (ubs *UpvoteBroadcasterService) Start(sessionStore *session.Store) {
	go func() {
		const broadcastInterval = 500 * time.Millisecond
		ticker := time.NewTicker(broadcastInterval)

		for {
			select {
			case playlistIDUpvoteMessage := <-ubs.broadcastQueue:
				log.Printf("adding new upvote to playlist %s to upvoted playlist list", playlistIDUpvoteMessage.PlaylistID)
				ubs.lock.Lock()
				ubs.playlistsBeingUpvoted = append(ubs.playlistsBeingUpvoted, playlistIDUpvoteMessage)
				ubs.lock.Unlock()
			case <-ticker.C:
				if len(ubs.playlistsBeingUpvoted) == 0 {
					continue
				}

				ubs.lock.Lock()
				playlistsBeingUpvoted := ubs.playlistsBeingUpvoted
				ubs.playlistsBeingUpvoted = []*domain.PlaylistIDUpvoteMessage{}
				ubs.lock.Unlock()

				sessions := sessionStore.GetAllUserSessions()

				if len(sessions) == 0 {
					continue
				}

				log.Printf("broadcasting upvotes to %d sessions to %d playlists", len(sessions), len(playlistsBeingUpvoted))

				for _, userSession := range sessions {
					for _, playlistID := range userSession.PlaylistIDsToListenToUpvoteChanges {
						i, found := sort.Find(len(playlistsBeingUpvoted), func(j int) int {
							return strings.Compare(string(playlistID), string(playlistsBeingUpvoted[j].PlaylistID))
						})

						if found {
							log.Printf("found upvote for playlist %s that user was interested in, sending to user %s", playlistID, userSession.ID)
							userSession.PlaylistIDUpvoteChannel <- playlistsBeingUpvoted[i]
						}
					}
				}
			}
		}
	}()
}

func (ubs *UpvoteBroadcasterService) Notify(playlistIDUpvoteMessage *domain.PlaylistIDUpvoteMessage) {
	select {
	case ubs.broadcastQueue <- playlistIDUpvoteMessage:
	default:
		log.Printf("upvote broadcaster queue full, dropping message")
	}
}
