package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metarang/training-service/internal/repository"
)

func TestVideoRepository_GetVideoBySlug_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id, video_sub_category_id").WithArgs("missing").WillReturnError(sql.ErrNoRows)

	r := repository.NewVideoRepository(db)
	v, err := r.GetVideoBySlug(context.Background(), "missing")
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Fatal("expected nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestVideoRepository_GetVideoBySlug_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "video_sub_category_id", "title", "slug", "description", "fileName", "creator_code", "image", "created_at", "updated_at"}).
		AddRow(uint64(1), uint64(2), "title", "my-slug", "d", "f.mp4", "c", "img.jpg", now, now)
	mock.ExpectQuery("SELECT id, video_sub_category_id").WithArgs("my-slug").WillReturnRows(rows)

	r := repository.NewVideoRepository(db)
	v, err := r.GetVideoBySlug(context.Background(), "my-slug")
	if err != nil || v == nil || v.Title != "title" {
		t.Fatalf("v=%+v err=%v", v, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestVideoRepository_AddInteraction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO interactions").WillReturnResult(sqlmock.NewResult(1, 1))

	r := repository.NewVideoRepository(db)
	if err := r.AddInteraction(context.Background(), 1, 2, true, "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestVideoRepository_IncrementView(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO views").WillReturnResult(sqlmock.NewResult(1, 1))

	r := repository.NewVideoRepository(db)
	if err := r.IncrementView(context.Background(), 9, "10.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
