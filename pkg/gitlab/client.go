package gitlab

import (
	"errors"

	gitlab "github.com/xanzy/go-gitlab"
)

type Client struct {
	note NoteServices
	mr   MergeRewuestsService
}

type ParamNew struct {
	Token         string
	GitLabBaseURL string
}

func New(param *ParamNew) (*Client, error) {
	client := &Client{}

	if param.Token == "" {
		return &Client{}, errors.New("gitlab token is missing")
	}

	gl, err := gitlab.NewClient(param.Token)
	if err != nil {
		return client, errors.New("failed to create a new gitlab api client")
	}

	if baseURL := param.GitLabBaseURL; baseURL != "" {
		gl, err = gitlab.NewClient(param.Token, gitlab.WithBaseURL(baseURL))
		if err != nil {
			return &Client{}, errors.New("failed to create a new gitlab api client")
		}
	}

	client.note = gl.Notes
	client.mr = gl.MergeRequests

	return client, nil
}

type NoteServices interface {
	CreateMergeRequestNote(pid interface{}, mergeRequest int, opt *gitlab.CreateMergeRequestNoteOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error)
	UpdateMergeRequestNote(pid interface{}, mergeRequest, note int, opt *gitlab.UpdateMergeRequestNoteOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error)
	ListMergeRequestNotes(pid interface{}, mergeRequest int, opt *gitlab.ListMergeRequestNotesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Note, *gitlab.Response, error)
}
type MergeRewuestsService interface{}
