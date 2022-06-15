package schema

import (
	"context"
	"errors"

	"github.com/firstcontributions/backend/internal/gateway/session"
	"github.com/firstcontributions/backend/internal/storemanager"
)

func (m *Resolver) CreateComment(
	ctx context.Context,
	args struct {
		Comment *CreateCommentInput
	},
) (*Comment, error) {
	session := session.FromContext(ctx)
	if session == nil {
		return nil, errors.New("unauthorized")
	}

	commentModelInput, err := args.Comment.ToModel()
	if err != nil {
		return nil, err
	}
	commentModelInput.CreatedBy = session.UserID()

	comment, err := storemanager.FromContext(ctx).StoriesStore.CreateComment(ctx, commentModelInput)
	if err != nil {
		return nil, err
	}
	return NewComment(comment), nil
}
func (m *Resolver) UpdateComment(
	ctx context.Context,
	args struct {
		Comment *UpdateCommentInput
	},
) (*Comment, error) {
	store := storemanager.FromContext(ctx)

	id, err := ParseGraphqlID(args.Comment.ID)
	if err != nil {
		return nil, err
	}
	if err := store.StoriesStore.UpdateComment(ctx, id.ID, args.Comment.ToModel()); err != nil {
		return nil, err
	}
	comment, err := store.StoriesStore.GetCommentByID(ctx, id.ID)
	if err != nil {
		return nil, err
	}
	return NewComment(comment), nil
}
