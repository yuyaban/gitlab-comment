package gitlab

import (
	"context"
	"fmt"
	"io"
	"strconv"
)

type Mock struct {
	Stderr   io.Writer
	Silent   bool
	Login    string
	MRNumber int
}

func (mock *Mock) CreateComment(ctx context.Context, note *Note) error {
	if mock.Silent {
		return nil
	}
	msg := "[gitlab-comment][DRYRUN] Comment to " + note.Org + "/" + note.Repo + " sha1:" + note.SHA1
	if note.MRNumber != 0 {
		msg += " MR:" + strconv.Itoa(note.MRNumber)
	}
	fmt.Fprintln(mock.Stderr, msg+"\n[gitlab-comment][DRYRUN] "+note.Body)
	return nil
}

func (mock *Mock) HideComment(ctx context.Context, nodeID int) error {
	return nil
}

func (mock *Mock) ListNote(ctx context.Context, mr *MergeRequest) ([]*Note, error) {
	return nil, nil
}

func (mock *Mock) GetAuthenticatedUser(ctx context.Context) (string, error) {
	return mock.Login, nil
}

func (mock *Mock) MRNumberWithSHA(ctx context.Context, owner, repo, sha string) (int, error) {
	return mock.MRNumber, nil
}
