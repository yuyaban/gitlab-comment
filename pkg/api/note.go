package api

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/github-comment-metadata/metadata"
	"github.com/yuyaban/gitlab-comment/pkg/gitlab"
)

// Gitlab is API to post a comment to GitHub
type Gitlab interface {
	CreateComment(note *gitlab.Note) error
	ListNote(mr *gitlab.MergeRequest) ([]*gitlab.Note, error)
	HideComment(nodeID int) error
	MRNumberWithSHA(owner, repo, sha string) (int, error)
}

type NoteController struct {
	Gitlab   Gitlab
	Expr     Expr
	Getenv   func(string) string
	Platform Platform
}

func (ctrl *NoteController) Post(ctx context.Context, note *gitlab.Note, hiddenParam map[string]interface{}) error {
	if err := ctrl.Gitlab.CreateComment(note); err != nil {
		return fmt.Errorf("send a comment: %w", err)
	}
	return nil
}

func extractMetaFromComment(body string, data *map[string]interface{}) bool {
	f, _ := metadata.Extract(body, data)
	return f
}

func (ctrl *NoteController) complementMetaData(data map[string]interface{}) {
	if data == nil {
		return
	}
	if ctrl.Platform == nil {
		return
	}
	_ = metadata.SetCIEnv(ctrl.Platform.CI(), ctrl.Getenv, data)
}

func (ctrl *NoteController) getEmbeddedComment(data map[string]interface{}) (string, error) {
	ctrl.complementMetaData(data)
	return metadata.Convert(data) //nolint:wrapcheck
}
