package util

import (
	"context"
	"os"
	"os/exec"

	"gbenson.net/go/logger"
)

func ExecSudo(ctx context.Context, name string, args ...string) error {
	if os.Getuid() == 0 {
		return Exec(ctx, name, args...)
	}
	return Exec(ctx, "sudo", append([]string{name}, args...)...)
}

func Exec(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	logger.Ctx(ctx).Debug().Stringer("cmd", cmd).Msg("Exec")
	return cmd.Run()
}
