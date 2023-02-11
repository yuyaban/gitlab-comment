package gitlab

import (
	"context"
	"fmt"

	gitlab "github.com/xanzy/go-gitlab"
)

type Note struct {
	ID             int
	MRNumber       int
	Org            string
	Repo           string
	NoteID         int
	Body           string
	BodyForTooLong string
	SHA1           string
	HideOldComment string
	Vars           map[string]interface{}
	TemplateKey    string
}

func (client *Client) sendMRComment(ctx context.Context, note *Note, body string) error {
	if note.NoteID != 0 {
		if _, _, err := client.note.UpdateMergeRequestNote(
			fmt.Sprintf("%s/%s", note.Org, note.Repo),
			note.MRNumber,
			note.NoteID,
			&gitlab.UpdateMergeRequestNoteOptions{Body: gitlab.String(body)},
		); err != nil {
			return fmt.Errorf("edit a merge request note by Gitlab API: %w", err)
		}
		return nil
	}
	if _, _, err := client.note.CreateMergeRequestNote(
		fmt.Sprintf("%s/%s", note.Org, note.Repo),
		note.MRNumber,
		&gitlab.CreateMergeRequestNoteOptions{Body: gitlab.String(body)},
	); err != nil {
		return fmt.Errorf("create a note to merge request by Gitlab API: %w", err)
	}
	return nil
}

func (client *Client) createComment(ctx context.Context, note *Note, tooLong bool) error {
	body := note.Body
	if tooLong {
		body = note.BodyForTooLong
	}
	if note.MRNumber != 0 {
		return client.sendMRComment(ctx, note, body)
	}
	return fmt.Errorf("not yet Support sendIssueComment method")
	// return client.sendCommitComment(ctx, note, body)
}

func (client *Client) CreateComment(ctx context.Context, note *Note) error {
	return client.createComment(ctx, note, len(note.Body) > 65536) //nolint:gomnd
}
