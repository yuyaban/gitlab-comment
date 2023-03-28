package gitlab

import (
	"fmt"
)

func (client *Client) MRNumberWithSHA(owner, repo, sha string) (int, error) {
	mrList, _, err := client.commit.ListMergeRequestsByCommit(
		fmt.Sprintf("%s/%s", owner, repo),
		sha,
	)
	if err != nil {
		return 0, err
	}

	if len(mrList) == 0 {
		return 0, fmt.Errorf("sha is not associated with MR")
	}

	return mrList[0].IID, nil
}
