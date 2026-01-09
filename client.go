package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/Khan/genqlient/graphql"
)

type cliClient struct {
	ghPath string
}

var _ graphql.Client = cliClient{}

func newCLIClient(ghPath string) cliClient {
	return cliClient{ghPath: ghPath}
}

// MakeRequest implements [graphql.Client].
func (c cliClient) MakeRequest(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
	body := struct {
		Query     string `json:"query"`
		Variables any    `json:"variables,omitempty"`
	}{
		Query:     req.Query,
		Variables: req.Variables,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	cmd := exec.CommandContext(ctx, c.ghPath, "api", "graphql", "--input", "-")
	cmd.Stdin = bytes.NewReader(bodyBytes)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("gh command failed: %w, stderr: %s", err, exitErr.Stderr)
		}
		return fmt.Errorf("gh command failed: %w", err)
	}

	if err := json.Unmarshal(output, resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
