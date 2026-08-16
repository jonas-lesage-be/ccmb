package media

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FrameType represents the type of media file based on the number of frames it contains.
type FrameType int

const (
	// UnknownFrameType indicates that the frame type could not be determined.
	UnknownFrameType FrameType = iota
	// SingleFrameType indicates that the media file contains a single frame.
	SingleFrameType
	// MultiFrameType indicates that the media file contains multiple frames.
	MultiFrameType
)

// MainType extracts the main type from a MIME type string.
func MainType(mtype string) string {
	mainType, _, found := strings.Cut(mtype, "/")

	if !found {
		return ""
	}

	return mainType
}

// ProbeFrameType uses ffprobe to determine the frame type of the media file at path.
// It only decodes the first two frames (via -read_intervals),
// which is enough to determine the answer without a full decode.
func ProbeFrameType(ctx context.Context, path string, timeout time.Duration) (FrameType, error) {
	output, err := runFFprobeCommand(ctx, path, timeout)
	if err != nil {
		return UnknownFrameType, err
	}

	frameCount, err := numberOfReadFrames(output)
	if err != nil {
		return UnknownFrameType, err
	}

	switch {
	case frameCount == 1:
		return SingleFrameType, nil
	case frameCount > 1:
		return MultiFrameType, nil
	default:
		return UnknownFrameType, nil
	}
}

func numberOfReadFrames(output []byte) (int, error) {
	cleanOutput := strings.TrimSpace(string(output))
	if cleanOutput == "" {
		return 0, nil
	}

	frameCount, err := strconv.Atoi(cleanOutput)
	if err != nil {
		return 0, fmt.Errorf("failed to convert frame count to integer: %w", err)
	}

	return frameCount, nil
}
