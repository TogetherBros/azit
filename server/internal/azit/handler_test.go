package azit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/db"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/room"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

// updateEnv 는 아지트 하나에 만든 사람(owner)과 일반 멤버(plain)가 있는 테스트 서버.
type updateEnv struct {
	app     *fiber.App
	azitID  int64
	owner   string // 토큰
	plain   string
	ownerID int64
}

func newUpdateEnv(t *testing.T) updateEnv {
	t.Helper()

	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("DB 열기 실패: %v", err)
	}
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() }) // 윈도우는 열린 파일을 못 지운다

	must(t, conn.AutoMigrate(&account.Account{}, &Azit{}, &member.Member{}, &room.Room{}))

	owner := account.Account{Provider: account.ProviderGoogle, Subject: "o", Email: "o@test.com", Nickname: "주인"}
	plain := account.Account{Provider: account.ProviderGoogle, Subject: "p", Email: "p@test.com", Nickname: "멤버"}
	must(t, conn.Create(&owner).Error)
	must(t, conn.Create(&plain).Error)

	repo := NewRepository(conn)
	a, err := repo.Create(t.Context(), CreateInput{MasterID: owner.ID, Nickname: "주인", Name: "의성", Color: "#D4FF3A"})
	must(t, err)
	must(t, conn.Create(&member.Member{AzitID: a.AzitID, AccountID: plain.ID, NickName: "멤버", Role: member.RoleMember}).Error)

	r2, err := storage.NewR2("test", "key", "secret", "bucket")
	must(t, err)
	memberRepo := member.NewRepository(conn)
	h := NewHandler(account.NewRepository(conn), memberRepo, repo, r2, nil, nil, nil, nil)

	p := auth.NewTokenProvider("test-secret", time.Minute, time.Hour)
	ot, err := p.IssueAccessToken(owner.ID)
	must(t, err)
	pt, err := p.IssueAccessToken(plain.ID)
	must(t, err)

	app := fiber.New(fiber.Config{ErrorHandler: httperr.Handler})
	h.Routes(app.Group("/v1/azits", auth.RequireAuth(p)), member.RequireMember(memberRepo))

	return updateEnv{app: app, azitID: a.AzitID, owner: ot, plain: pt, ownerID: owner.ID}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func (e updateEnv) patch(t *testing.T, token string, body string) (int, proto.AzitResponse) {
	t.Helper()
	r := httptest.NewRequest("PATCH", fmt.Sprintf("/v1/azits/%d", e.azitID), bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	res, err := e.app.Test(r)
	must(t, err)
	var out proto.AzitResponse
	if res.StatusCode == 200 {
		must(t, json.NewDecoder(res.Body).Decode(&out))
	}
	return res.StatusCode, out
}

func TestUpdate_OnlySentFields(t *testing.T) {
	e := newUpdateEnv(t)

	status, a := e.patch(t, e.owner, `{"name":"  의성 낚시회 ","description":"낚시하다 강아지 낚은 사람들"}`)
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if a.Name != "의성 낚시회" || a.Description != "낚시하다 강아지 낚은 사람들" {
		t.Errorf("name, description = %q, %q", a.Name, a.Description)
	}
	if a.Color != "#D4FF3A" {
		t.Errorf("안 보낸 색이 바뀜: %q", a.Color)
	}
	if a.MyRole != "super_admin" || a.MemberCount != 2 {
		t.Errorf("my_role, member_count = %q, %d", a.MyRole, a.MemberCount)
	}

	// 소개 지우기 + 색 바꾸기 (소문자로 와도 됨)
	status, a = e.patch(t, e.owner, `{"description":"","color":"#ff5fa8"}`)
	if status != 200 || a.Description != "" || a.Color != "#FF5FA8" || a.Name != "의성 낚시회" {
		t.Errorf("status %d, got %+v", status, a)
	}
}

func TestUpdate_Rejects(t *testing.T) {
	e := newUpdateEnv(t)

	cases := []struct {
		name  string
		token string
		body  string
		want  int
	}{
		{"일반 멤버", e.plain, `{"name":"바꿈"}`, 403},
		{"빈 이름", e.owner, `{"name":"   "}`, 400},
		{"긴 이름", e.owner, `{"name":"123456789012345678901"}`, 400},
		{"목록에 없는 색", e.owner, `{"color":"#000000"}`, 400},
		{"긴 소개", e.owner, `{"description":"12345678901234567890123456789012345678901"}`, 400},
		{"남의 배너", e.owner, `{"banner_key":"azit-banners/999/X.webp"}`, 400},
		{"배너 폴더가 아닌 키", e.owner, fmt.Sprintf(`{"banner_key":"avatars/%d/X.webp"}`, e.ownerID), 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if status, _ := e.patch(t, tc.token, tc.body); status != tc.want {
				t.Errorf("status = %d, want %d", status, tc.want)
			}
		})
	}
}
