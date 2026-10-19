package main

import (
	"context"
	"errors"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/azit"
	"github.com/TogetherBros/azit/server/internal/changelog"
	"github.com/TogetherBros/azit/server/internal/live"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/meme"
	"github.com/TogetherBros/azit/server/internal/room"
)

// registerUndoers 는 변경 기록 종류마다 "되돌리는 법" 을 꽂는다.
// changelog 는 다른 패키지를 모르고, 여기(main)가 repo 들을 다 알아서 이어 준다.
// 되돌린 뒤에는 그 목록이 바뀌었다고 멤버들에게 실시간으로 알린다.
// 방 만들기(room.create)는 방을 지우지 않기로 해서 되돌리기가 없다.
func registerUndoers(logs *changelog.Repository, lv *live.Live, azits *azit.Repository, rooms *room.Repository, members *member.Repository, memes *meme.Repository) {
	logs.Handle(changelog.KindAzitInfo, func(ctx context.Context, e changelog.Entry, _ int64) error {
		var before map[string]any // 바뀐 칸만: name, color, description, banner_key
		if err := changelog.DecodeBefore(e, &before); err != nil {
			return err
		}
		if err := azits.Update(ctx, e.AzitID, before); err != nil {
			return err
		}
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedInfo)
		return nil
	})

	logs.Handle(changelog.KindRoomRename, func(ctx context.Context, e changelog.Entry, _ int64) error {
		var before struct{ Name string }
		if err := changelog.DecodeBefore(e, &before); err != nil {
			return err
		}
		r, err := rooms.FindByID(ctx, e.TargetID)
		if err != nil || r.AzitID != e.AzitID {
			return errors.New("방이 없어요")
		}
		if err := rooms.Rename(ctx, e.TargetID, before.Name); err != nil {
			return err
		}
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedRooms)
		return nil
	})

	logs.Handle(changelog.KindRoomOrder, func(ctx context.Context, e changelog.Entry, _ int64) error {
		var before struct {
			RoomIDs []int64 `json:"room_ids"`
		}
		if err := changelog.DecodeBefore(e, &before); err != nil {
			return err
		}
		if err := rooms.Reorder(ctx, e.AzitID, before.RoomIDs); err != nil {
			if errors.Is(err, room.ErrBadOrder) {
				return errors.New("그 뒤로 방이 생겨서 예전 순서로 못 돌려요")
			}
			return err
		}
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedRooms)
		return nil
	})

	logs.Handle(changelog.KindRoomDelete, func(ctx context.Context, e changelog.Entry, _ int64) error {
		if err := rooms.Restore(ctx, e.AzitID, e.TargetID); err != nil {
			if errors.Is(err, room.ErrNotFound) {
				return errors.New("이미 다시 살린 방이에요")
			}
			return err
		}
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedRooms)
		return nil
	})

	logs.Handle(changelog.KindMemberRole, func(ctx context.Context, e changelog.Entry, _ int64) error {
		var before struct{ Role string }
		if err := changelog.DecodeBefore(e, &before); err != nil {
			return err
		}
		if _, err := members.FindActive(ctx, e.AzitID, e.TargetID); err != nil {
			return errors.New("이제 이 아지트 멤버가 아니에요")
		}
		if err := members.UpdateRole(ctx, e.AzitID, e.TargetID, member.Role(before.Role)); err != nil {
			return err
		}
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedMembers)
		return nil
	})

	logs.Handle(changelog.KindMemberKick, func(ctx context.Context, e changelog.Entry, _ int64) error {
		if err := members.Restore(ctx, e.AzitID, e.TargetID); err != nil {
			return err
		}
		lv.AzitChangedFor(ctx, e.AzitID, proto.ChangedMembers, e.TargetID) // 다시 들어간 사람 앱의 아지트 목록도
		return nil
	})

	logs.Handle(changelog.KindMemberUnban, func(ctx context.Context, e changelog.Entry, _ int64) error {
		var before member.Ban
		if err := changelog.DecodeBefore(e, &before); err != nil {
			return err
		}
		return members.Reban(ctx, before)
	})

	logs.Handle(changelog.KindMemeCreate, func(ctx context.Context, e changelog.Entry, _ int64) error {
		if _, err := memes.FindActive(ctx, e.AzitID, e.TargetID); err != nil {
			return errors.New("이미 빠진 짤이에요")
		}
		if err := memes.SoftDelete(ctx, e.TargetID); err != nil {
			return err
		}
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedMemes)
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedReacts)
		return nil
	})

	logs.Handle(changelog.KindMemeDelete, func(ctx context.Context, e changelog.Entry, _ int64) error {
		if err := memes.Undelete(ctx, e.AzitID, e.TargetID); err != nil {
			return err
		}
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedMemes)
		return nil
	})

	logs.Handle(changelog.KindReacts, func(ctx context.Context, e changelog.Entry, _ int64) error {
		var before []proto.ReactItem
		if err := changelog.DecodeBefore(e, &before); err != nil {
			return err
		}
		reacts := make([]meme.React, 0, len(before))
		for _, item := range before {
			r := meme.React{Text: item.Text}
			if item.MemeID > 0 {
				if _, err := memes.FindActive(ctx, e.AzitID, item.MemeID); err != nil {
					continue // 그 사이 빠진 짤 칸은 건너뛴다
				}
				id := item.MemeID
				r.MemeID = &id
			}
			reacts = append(reacts, r)
		}
		if err := memes.ReplaceReacts(ctx, e.AzitID, reacts); err != nil {
			return err
		}
		lv.AzitChanged(ctx, e.AzitID, proto.ChangedReacts)
		return nil
	})
}
