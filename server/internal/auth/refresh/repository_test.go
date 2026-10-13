package refresh

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/TogetherBros/azit/server/internal/db"
)

// newTestRepository 는 테스트마다 새 임시 DB 파일로 Repository 를 만든다.
func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("DB 열기 실패: %v", err)
	}
	if err := conn.AutoMigrate(&Token{}); err != nil {
		t.Fatalf("AutoMigrate 실패: %v", err)
	}
	return NewRepository(conn)
}

func TestConsume_ReturnsAccountID(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	if err := repo.Save(ctx, 14, "hash-a", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	got, err := repo.Consume(ctx, "hash-a")
	if err != nil {
		t.Fatalf("Consume 실패: %v", err)
	}
	if got != 14 {
		t.Errorf("got %d, want 14", got)
	}
}

// 한 번 쓴 RT 는 다시 못 쓴다 (회전)
func TestConsume_OnlyOnce(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	repo.Save(ctx, 14, "hash-a", time.Now().Add(time.Hour))
	repo.Consume(ctx, "hash-a")

	if _, err := repo.Consume(ctx, "hash-a"); !errors.Is(err, ErrInvalid) {
		t.Errorf("두 번째 Consume: got %v, want ErrInvalid", err)
	}
}

func TestConsume_Expired(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	repo.Save(ctx, 14, "hash-a", time.Now().Add(-time.Minute)) // 이미 만료

	if _, err := repo.Consume(ctx, "hash-a"); !errors.Is(err, ErrInvalid) {
		t.Errorf("got %v, want ErrInvalid", err)
	}
}

func TestConsume_Unknown(t *testing.T) {
	repo := newTestRepository(t)

	if _, err := repo.Consume(context.Background(), "없는-hash"); !errors.Is(err, ErrInvalid) {
		t.Errorf("got %v, want ErrInvalid", err)
	}
}

// 로그아웃(Delete) 한 RT 는 못 쓴다
func TestDelete(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	repo.Save(ctx, 14, "hash-a", time.Now().Add(time.Hour))
	if err := repo.Delete(ctx, "hash-a"); err != nil {
		t.Fatalf("Delete 실패: %v", err)
	}

	if _, err := repo.Consume(ctx, "hash-a"); !errors.Is(err, ErrInvalid) {
		t.Errorf("got %v, want ErrInvalid", err)
	}
}
