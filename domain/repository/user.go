package repository

import (
	"context"
	"github.com/goodfoodcesi/auth-api/infrastructure/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	Create(ctx context.Context, user *db.User) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*db.User, error)
	Update(ctx context.Context, user *db.User) error
}

// Definition for a binary tree node.
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func removeDuplicates(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	d := 0
	var m = make(map[int]bool)
	for i, value := range nums {
		if m[value] == false {
			nums[i] = value
			m[value] = true
			d++
		}
	}
	return d
}
