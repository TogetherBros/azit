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

	a := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Nickname: "문경"}
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

	first := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Nickname: "문경"}
	if err := repo.Create(ctx, &first); err != nil {
		t.Fatalf("첫 Create 실패: %v", err)
	}

	second := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "b@test.com", Nickname: "다른 이름"}
	if err := repo.Create(ctx, &second); err == nil {
		t.Error("같은 provider+subject 인데 두 번째 Create 가 성공함")
	}
}

func TestIsNameTaken(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	name := "moongyeong"
	a := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Nickname: "문경", Name: &name}
	if err := repo.Create(ctx, &a); err != nil {
		t.Fatalf("Create 실패: %v", err)
	}
	if taken, err := repo.IsNameTaken(ctx, "moongyeong"); err != nil || !taken {
		t.Errorf("쓰는 이름인데 taken=%v err=%v", taken, err)
	}
	if taken, err := repo.IsNameTaken(ctx, "nobody_here"); err != nil || taken {
		t.Errorf("안 쓰는 이름인데 taken=%v err=%v", taken, err)
	}
}

// 기본 후보가 이미 있으면 뒤에 숫자 4자리를 붙인다
func TestSuggestName(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	got, err := repo.SuggestName(ctx, "Moon.Gyeong@gmail.com")
	if err != nil || got != "moongyeong" {
		t.Fatalf("got %q err=%v, want moongyeong", got, err)
	}

	a := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Nickname: "문경", Name: &got}
	if err := repo.Create(ctx, &a); err != nil {
		t.Fatalf("Create 실패: %v", err)
	}
	second, err := repo.SuggestName(ctx, "moongyeong@other.com")
	if err != nil {
		t.Fatal(err)
	}
	if second == got || len(second) != len(got)+4 || ValidateName(second) != nil {
		t.Errorf("겹칠 때 추천값이 이상함: %q", second)
	}
}

func TestUpdateProfile(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	a := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Nickname: "정문경", Bio: "예전 소개"}
	if err := repo.Create(ctx, &a); err != nil {
		t.Fatalf("Create 실패: %v", err)
	}

	err := repo.UpdateProfile(ctx, a.ID, ProfileUpdate{Name: "moongyeong", Nickname: "문경", Bio: "", Color: "#7BC4FF"})
	if err != nil {
		t.Fatalf("UpdateProfile 실패: %v", err)
	}
	got, _ := repo.FindByID(ctx, a.ID)
	if got.Name == nil || *got.Name != "moongyeong" || got.Nickname != "문경" || got.Color != "#7BC4FF" {
		t.Errorf("저장이 이상함: %+v", got)
	}
	if got.Bio != "" {
		t.Errorf("빈 소개도 저장돼야 하는데 %q", got.Bio)
	}
}

func TestUpdateProfile_NameTaken(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	first := Account{Provider: ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Nickname: "문경"}
	second := Account{Provider: ProviderGoogle, Subject: "sub-2", Email: "b@test.com", Nickname: "백호"}
	repo.Create(ctx, &first)
	repo.Create(ctx, &second)

	if err := repo.UpdateProfile(ctx, first.ID, ProfileUpdate{Name: "samename", Nickname: "문경", Color: "#D4FF3A"}); err != nil {
		t.Fatalf("첫 저장 실패: %v", err)
	}
	err := repo.UpdateProfile(ctx, second.ID, ProfileUpdate{Name: "samename", Nickname: "백호", Color: "#D4FF3A"})
	if !errors.Is(err, ErrNameTaken) {
		t.Errorf("got %v, want ErrNameTaken", err)
	}
}
