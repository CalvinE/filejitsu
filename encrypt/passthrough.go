package encrypt

import (
	"github.com/calvine/filejitsu/util"
	"golang.org/x/exp/slog"
)

func Passthrough(logger *slog.Logger, params Params) error {
	err := util.ProcessStreams(logger, params.Input, params.Output)
	if err != nil {
		logger.Error("failed to passthrough data", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("done passing through input")
	return nil
}
