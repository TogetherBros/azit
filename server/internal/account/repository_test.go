package account

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/TogetherBros/azit/server/internal/db"
)

// newTestRepository 는 테스트마다 새 임시 DB 파일로 Repository 를 만든다.
func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("DB 열기 실패: %v", err)
	}
	if err := conn.AutoMigrate(&Account{}); err != nil {
		t.Fatalf("AutoMigrate 실패: %v", err)
	}
	return NewRepository(conn)
}

func TestCreate_ThenFind(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	a := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Name: "문경"}
	if err := repo.Create(ctx, &a); err != nil {
		t.Fatalf("Create 실패: %v", err)
	}
	if a.ID == 0 {
		t.Fatal("Create 후 ID 가 채워지지 않음")
	}

	got, err := repo.FindByProviderSubject(ctx, ProviderGoogle, "sub-1")
	if err != nil {
		t.Fatalf("Find 실패: %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("got ID %d, want %d", got.ID, a.ID)
	}
}

func TestFind_Unknown(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.FindByProviderSubject(context.Background(), ProviderGoogle, "없는-sub")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

// 같은 provider+subject 는 두 번 못 만든다 (복합 유니크)
func TestCreate_Duplicate(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	first := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Name: "문경"}
	if err := repo.Create(ctx, &first); err != nil {
		t.Fatalf("첫 Create 실패: %v", err)
	}

	second := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "b@test.com", Name: "다른 이름"}
	if err := repo.Create(ctx, &second); err == nil {
		t.Error("같은 provider+subject 인데 두 번째 Create 가 성공함")
	}
}
