package gitlab

import (
	"context"
	"fmt"
)

func (client *Client) MRNumberWithSHA(ctx context.Context, owner, repo, sha string) (int, error) {
	return 0, fmt.Errorf("not yet Supported MRNumberWithSHA method")
	// prs, _, err := client.mr.ListMergeRequestsWithCommit(ctx, owner, repo, sha, &gitlab.MergeRequestListOptions{
	// 	State: "all",
	// 	Sort:  "updated",
	// 	ListOptions: gitlab.ListOptions{
	// 		PerPage: 1,
	// 	},
	// })
	// if err != nil {
	// 	return 0, fmt.Errorf("list associated merge requests: %w", err)
	// }
	// if len(prs) == 0 {
	// 	return 0, errors.New("associated merge request isn't found")
	// }
	// return prs[0].GetNumber(), nil
}
