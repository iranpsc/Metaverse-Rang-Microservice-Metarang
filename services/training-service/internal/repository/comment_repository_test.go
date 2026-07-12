package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metarang/training-service/internal/repository"
)

func TestCommentRepository_GetComments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").WithArgs(uint64(10)).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(int32(1)))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "parent_id", "commentable_type", "commentable_id", "content", "created_at", "updated_at"}).
		AddRow(uint64(1), uint64(2), nil, "App\\Models\\Video", uint64(10), "hi", now, now)
	mock.ExpectQuery("SELECT id, user_id, parent_id").WithArgs(uint64(10), int32(10), int32(0)).WillReturnRows(rows)

	r := repository.NewCommentRepository(db)
	list, total, err := r.GetComments(context.Background(), 10, 1, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("err=%v total=%d n=%d", err, total, len(list))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCommentRepository_AddCommentInteraction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO interactions").WillReturnResult(sqlmock.NewResult(1, 1))

	r := repository.NewCommentRepository(db)
	if err := r.AddCommentInteraction(context.Background(), 5, 6, true, "ip"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
