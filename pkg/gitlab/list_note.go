package gitlab

import (
	"fmt"

	"github.com/jinzhu/copier"
	gitlab "github.com/xanzy/go-gitlab"
)

type MergeRequest struct {
	MRNumber int
	Org      string
	Repo     string
}

// func (client *Client) listIssueComment(ctx context.Context, pr *PullRequest) ([]*IssueComment, error) { //nolint:dupl
// 	// https://github.com/shurcooL/githubv4#pagination
// 	var q struct {
// 		Repository struct {
// 			Issue struct {
// 				Comments struct {
// 					Nodes    []*IssueComment
// 					PageInfo struct {
// 						EndCursor   githubv4.String
// 						HasNextPage bool
// 					}
// 				} `graphql:"comments(first: 100, after: $commentsCursor)"` // 100 per page.
// 			} `graphql:"issue(number: $issueNumber)"`
// 		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
// 	}
// 	variables := map[string]interface{}{
// 		"repositoryOwner": githubv4.String(pr.Org),
// 		"repositoryName":  githubv4.String(pr.Repo),
// 		"issueNumber":     githubv4.Int(pr.PRNumber),
// 		"commentsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
// 	}

// 	var allComments []*IssueComment
// 	for {
// 		if err := client.ghV4.Query(ctx, &q, variables); err != nil {
// 			return nil, fmt.Errorf("list issue comments by GitHub API: %w", err)
// 		}
// 		allComments = append(allComments, q.Repository.Issue.Comments.Nodes...)
// 		if !q.Repository.Issue.Comments.PageInfo.HasNextPage {
// 			break
// 		}
// 		variables["commentsCursor"] = githubv4.NewString(q.Repository.Issue.Comments.PageInfo.EndCursor)
// 	}
// 	return allComments, nil
// }

func (client *Client) listMRNote(mr *MergeRequest) ([]*Note, error) {
	notes, _, err := client.note.ListMergeRequestNotes(
		fmt.Sprintf("%s/%s", mr.Org, mr.Repo),
		mr.MRNumber,
		&gitlab.ListMergeRequestNotesOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("list Notes by GitLab API: %w", err)
	}

	var allNotes []*Note
	if err := copier.Copy(&allNotes, &notes); err != nil {
		return nil, fmt.Errorf("fetch list Notes: %w", err)
	}

	return allNotes, nil
}

func (client *Client) ListNote(mr *MergeRequest) ([]*Note, error) {
	notes, mrErr := client.listMRNote(mr)
	if mrErr == nil {
		return notes, nil
	}
	// notes, err := client.listIssueComment(ctx, pr)
	// if err == nil {
	// 	return notes, nil
	// }
	return nil, fmt.Errorf("get merge request notes: %w", mrErr)
}
