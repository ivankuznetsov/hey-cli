package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/basecamp/hey-cli/internal/output"
)

type postingActionCommand struct {
	cmd       *cobra.Command
	pastTense string
	kind      string
	runSingle func(context.Context, int64) error
}

func newPostingActionCommand(name, short, pastTense, agentNotes string, runSingle func(context.Context, int64) error, aliases ...string) *postingActionCommand {
	c := &postingActionCommand{
		pastTense: pastTense,
		runSingle: runSingle,
	}
	c.cmd = &cobra.Command{
		Use:     name + " <posting-id>",
		Aliases: aliases,
		Short:   short,
		Example: fmt.Sprintf("  hey %s 12345 --kind topic", name),
		Annotations: map[string]string{
			"agent_notes": agentNotes + " Pass --kind exactly as returned by hey box --json. HEY World posts are rejected before any email action is requested.",
		},
		RunE: c.run,
		Args: usageExactOneArg(),
	}
	c.cmd.Flags().StringVar(&c.kind, "kind", "", "Posting kind from hey box --json (required)")

	return c
}

func (c *postingActionCommand) run(cmd *cobra.Command, args []string) error {
	ids, err := parseIntArgs(args)
	if err != nil {
		return err
	}

	switch writer.EffectiveFormat() {
	case output.FormatIDs:
		return output.ErrUsage("--ids-only requires list data")
	case output.FormatCount:
		return output.ErrUsage("--count requires list data")
	default:
	}
	if err := c.validateKind(); err != nil {
		return err
	}
	if err := requireAuth(); err != nil {
		return err
	}

	if err := c.runSingle(cmd.Context(), ids[0]); err != nil {
		return convertSDKError(err)
	}

	summary := fmt.Sprintf("Posting %d %s", ids[0], c.pastTense)

	if writer.IsStyled() {
		fmt.Fprintln(cmd.OutOrStdout(), summary+".")
		return nil
	}

	return writeOK(nil, output.WithSummary(summary))
}

func (c *postingActionCommand) validateKind() error {
	kind := strings.TrimSpace(c.kind)
	if kind == "" {
		return output.ErrUsageHint(
			"--kind is required for posting actions",
			"Use the exact kind from `hey box <box> --json`, for example `--kind topic`.",
		)
	}
	if !strings.EqualFold(kind, "world/post") {
		return nil
	}

	if c.cmd.Name() == "trash" {
		return output.ErrUsageHint(
			"hey trash cannot move a HEY World post to Trash",
			"HEY World posts are published content. Deleting one requires a separate action and explicit confirmation.",
		)
	}
	return output.ErrUsageHint(
		fmt.Sprintf("hey %s cannot act on a HEY World post; it only works with email postings", c.cmd.Name()),
		"HEY World posts are published content and cannot be handled by email posting actions.",
	)
}

func newPaperTrailCommand() *postingActionCommand {
	return newPostingActionCommand(
		"paper-trail",
		"Move a posting to Paper Trail",
		"moved to Paper Trail",
		"State-changing action. Confirm the exact posting ID and obtain explicit user approval before moving the posting to Paper Trail.",
		func(ctx context.Context, id int64) error {
			return sdk.Postings().MoveToPaperTrail(ctx, id)
		},
		"papertrail",
		"trail",
	)
}

func newTrashCommand() *postingActionCommand {
	return newPostingActionCommand(
		"trash",
		"Move a posting to Trash",
		"moved to Trash",
		"State-changing action. Confirm the exact posting ID and obtain explicit user approval before moving the posting to Trash.",
		func(ctx context.Context, id int64) error {
			return sdk.Postings().MoveToTrash(ctx, id)
		},
	)
}

func newFeedCommand() *postingActionCommand {
	return newPostingActionCommand(
		"feed",
		"Move a posting to The Feed",
		"moved to The Feed",
		"State-changing action. Confirm the exact posting ID and obtain explicit user approval before moving the posting to The Feed.",
		func(ctx context.Context, id int64) error {
			return sdk.Postings().MoveToFeed(ctx, id)
		},
	)
}

func newSetAsideCommand() *postingActionCommand {
	return newPostingActionCommand(
		"set-aside",
		"Move a posting to Set Aside",
		"moved to Set Aside",
		"State-changing action. Confirm the exact posting ID and obtain explicit user approval before moving the posting to Set Aside.",
		func(ctx context.Context, id int64) error {
			return sdk.Postings().MoveToSetAside(ctx, id)
		},
		"aside",
	)
}

func newReplyLaterCommand() *postingActionCommand {
	return newPostingActionCommand(
		"reply-later",
		"Move a posting to Reply Later",
		"moved to Reply Later",
		"State-changing action. Confirm the exact posting ID and obtain explicit user approval before moving the posting to Reply Later.",
		func(ctx context.Context, id int64) error {
			return sdk.Postings().MoveToReplyLater(ctx, id)
		},
		"later",
	)
}

func newIgnoreCommand() *postingActionCommand {
	return newPostingActionCommand(
		"ignore",
		"Ignore a posting",
		"ignored",
		"State-changing action. Confirm the exact posting ID and obtain explicit user approval before ignoring the posting.",
		func(ctx context.Context, id int64) error {
			return sdk.Postings().Ignore(ctx, id)
		},
	)
}
