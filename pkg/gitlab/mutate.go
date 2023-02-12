package gitlab

import (
	"fmt"
)

func (client *Client) HideComment(nodeID int) error {
	return fmt.Errorf("maybe gitLab not Support hide Comment")
	// TBD: I'll make it when GitLab supports the hide option.
}
