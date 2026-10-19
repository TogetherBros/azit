// Package live 는 "무엇이 바뀌었다" 를 그 아지트 사람들에게 실시간으로 알리는 도우미다.
// 받는 사람을 고르는 일(멤버 목록 조회)과 허브로 넘기는 일을 한곳에 모아서,
// 방·아지트·멤버·짤 핸들러가 한 줄(AzitChanged)로 부를 수 있게 한다.
//
// 내용은 보내지 않는다. 앱은 신호를 받으면 그 목록을 HTTP 로 다시 받는다 (메시지 고침과 같은 방식).
package live

import (
	"context"
	"log"

	"github.com/TogetherBros/azit/proto"
)

// Publisher 는 허브 (hub.Hub).
type Publisher interface {
	Publish(accountIDs []int64, ev any) error
}

// Members 는 받을 사람을 고르는 쪽 (member.Repository).
type Members interface {
	ActiveIDs(ctx context.Context, azitID int64) ([]int64, error)
	CoMemberIDs(ctx context.Context, accountID int64) ([]int64, error)
}

type Live struct {
	pub     Publisher
	members Members
}

func New(pub Publisher, members Members) *Live {
	return &Live{pub: pub, members: members}
}

// AzitChanged 는 이 아지트 지금 멤버 전원에게 "what 이 바뀌었다" 를 보낸다. 저장(커밋)이 끝난 뒤에 부른다.
// l 이 nil 이면 아무것도 안 한다 (테스트에서 실시간이 필요 없을 때).
func (l *Live) AzitChanged(ctx context.Context, azitID int64, what string) {
	if l == nil {
		return
	}
	ids, err := l.members.ActiveIDs(ctx, azitID)
	if err != nil {
		log.Printf("실시간 배달 실패 (멤버 조회): %v", err)
		return
	}
	l.publish(ids, proto.Event{Type: proto.EventAzitChanged, Data: proto.AzitChange{AzitID: azitID, What: what}})
}

// AzitChangedFor 는 아지트에서 빠진 사람처럼, 멤버 목록에 없는 사람에게도 같이 알려야 할 때 쓴다.
func (l *Live) AzitChangedFor(ctx context.Context, azitID int64, what string, extra ...int64) {
	if l == nil {
		return
	}
	ids, err := l.members.ActiveIDs(ctx, azitID)
	if err != nil {
		log.Printf("실시간 배달 실패 (멤버 조회): %v", err)
		return
	}
	l.publish(append(ids, extra...), proto.Event{Type: proto.EventAzitChanged, Data: proto.AzitChange{AzitID: azitID, What: what}})
}

// Presence 는 accountID 가 앱을 켰다(online)/껐다 를 같은 아지트 사람들에게 알린다.
func (l *Live) Presence(ctx context.Context, accountID int64, online bool) {
	if l == nil {
		return
	}
	ids, err := l.members.CoMemberIDs(ctx, accountID)
	if err != nil {
		log.Printf("실시간 배달 실패 (접속 알림): %v", err)
		return
	}
	l.publish(ids, proto.Event{Type: proto.EventPresence, Data: proto.Presence{AccountID: accountID, Online: online}})
}

func (l *Live) publish(ids []int64, ev proto.Event) {
	if len(ids) == 0 {
		return
	}
	if err := l.pub.Publish(ids, ev); err != nil {
		log.Printf("실시간 배달 실패: %v", err)
	}
}
