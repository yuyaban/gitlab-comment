package gitlab

import (
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

func (mock *Mock) CreateComment(note *Note) error {
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

func (mock *Mock) HideComment(nodeID int) error {
	return nil
}

func (mock *Mock) ListNote(mr *MergeRequest) ([]*Note, error) {
	return nil, nil
}

func (mock *Mock) MRNumberWithSHA(owner, repo, sha string) (int, error) {
	return mock.MRNumber, nil
}
