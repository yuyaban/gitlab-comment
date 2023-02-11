package gitlab

import (
	"context"
	"fmt"
)

func (client *Client) HideComment(ctx context.Context, nodeID int) error {
	return fmt.Errorf("maybe gitLab not Support hide Comment")
	// TBD: I'll make it when Gitlab supports the hide option.
}
