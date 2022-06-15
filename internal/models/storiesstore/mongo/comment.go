package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/firstcontributions/backend/internal/models/storiesstore"
	"github.com/firstcontributions/backend/internal/models/utils"
	"github.com/firstcontributions/backend/pkg/cursor"
	"github.com/gokultp/go-mongoqb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *StoriesStore) CreateComment(ctx context.Context, comment *storiesstore.Comment) (*storiesstore.Comment, error) {
	now := time.Now()
	comment.TimeCreated = now
	comment.TimeUpdated = now
	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	comment.Id = uuid.String()
	if _, err := s.getCollection(CollectionComments).InsertOne(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *StoriesStore) GetCommentByID(ctx context.Context, id string) (*storiesstore.Comment, error) {
	qb := mongoqb.NewQueryBuilder().
		Eq("_id", id)
	var comment storiesstore.Comment
	if err := s.getCollection(CollectionComments).FindOne(ctx, qb.Build()).Decode(&comment); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (s *StoriesStore) GetComments(
	ctx context.Context,
	ids []string,
	story *storiesstore.Story,
	after *string,
	before *string,
	first *int64,
	last *int64,
) (
	[]*storiesstore.Comment,
	bool,
	bool,
	string,
	string,
	error,
) {
	qb := mongoqb.NewQueryBuilder()
	if len(ids) > 0 {
		qb.In("_id", ids)
	}
	if story != nil {
		qb.Eq("story_id", story.Id)
	}

	limit, order, cursorStr := utils.GetLimitAndSortOrderAndCursor(first, last, after, before)
	var c *cursor.Cursor
	if cursorStr != nil {
		c = cursor.FromString(*cursorStr)
		if c != nil {
			if order == 1 {
				qb.Lte("time_created", c.TimeStamp)
				qb.Lte("_id", c.ID)
			} else {
				qb.Gte("time_created", c.TimeStamp)
				qb.Gte("_id", c.ID)
			}
		}
	}
	sortOrder := utils.GetSortOrder(order)
	// incrementing limit by 2 to check if next, prev elements are present
	limit += 2
	options := &options.FindOptions{
		Limit: &limit,
		Sort:  sortOrder,
	}

	var firstCursor, lastCursor string
	var hasNextPage, hasPreviousPage bool

	var comments []*storiesstore.Comment
	mongoCursor, err := s.getCollection(CollectionComments).Find(ctx, qb.Build(), options)
	if err != nil {
		return nil, hasNextPage, hasPreviousPage, firstCursor, lastCursor, err
	}
	err = mongoCursor.All(ctx, &comments)
	if err != nil {
		return nil, hasNextPage, hasPreviousPage, firstCursor, lastCursor, err
	}
	count := len(comments)
	if count == 0 {
		return comments, hasNextPage, hasPreviousPage, firstCursor, lastCursor, nil
	}

	// check if the cursor element present, if yes that can be a prev elem
	if c != nil && comments[0].Id == c.ID {
		hasPreviousPage = true
		comments = comments[1:]
		count--
	}

	// check if actual limit +1 elements are there, if yes trim it to limit
	if count >= int(limit)-1 {
		hasNextPage = true
		comments = comments[:limit-2]
		count = len(comments)
	}

	if count > 0 {
		firstCursor = cursor.NewCursor(comments[0].Id, comments[0].TimeCreated).String()
		lastCursor = cursor.NewCursor(comments[count-1].Id, comments[count-1].TimeCreated).String()
	}
	if order < 0 {
		hasNextPage, hasPreviousPage = hasPreviousPage, hasNextPage
		firstCursor, lastCursor = lastCursor, firstCursor
	}
	return comments, hasNextPage, hasPreviousPage, firstCursor, lastCursor, nil
}
func (s *StoriesStore) UpdateComment(ctx context.Context, id string, commentUpdate *storiesstore.CommentUpdate) error {
	qb := mongoqb.NewQueryBuilder().
		Eq("_id", id)

	now := time.Now()
	commentUpdate.TimeUpdated = &now

	u := mongoqb.NewUpdateMap().
		SetFields(commentUpdate)

	um, err := u.BuildUpdate()
	if err != nil {
		return err
	}
	if _, err := s.getCollection(CollectionComments).UpdateOne(ctx, qb.Build(), um); err != nil {
		return err
	}
	return nil
}

func (s *StoriesStore) DeleteCommentByID(ctx context.Context, id string) error {
	qb := mongoqb.NewQueryBuilder().
		Eq("_id", id)
	if _, err := s.getCollection(CollectionComments).DeleteOne(ctx, qb.Build()); err != nil {
		return err
	}
	return nil
}
