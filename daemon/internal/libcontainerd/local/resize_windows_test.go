//go:build windows

package local

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/Microsoft/hcsshim"
	"github.com/containerd/log"
)

type resizeTestHCSProcess struct {
	width  uint16
	height uint16
}

var _ hcsshim.Process = (*resizeTestHCSProcess)(nil)

func (m *resizeTestHCSProcess) Pid() int { return 123 }
func (m *resizeTestHCSProcess) Kill() error { return nil }
func (m *resizeTestHCSProcess) Wait() error { return nil }
func (m *resizeTestHCSProcess) WaitTimeout(time.Duration) error { return nil }
func (m *resizeTestHCSProcess) ExitCode() (int, error) { return 0, nil }

func (m *resizeTestHCSProcess) ResizeConsole(width, height uint16) error {
	m.width = width
	m.height = height
	return nil
}

func (m *resizeTestHCSProcess) Stdio() (io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	return nil, nil, nil, nil
}

func (m *resizeTestHCSProcess) CloseStdin() error { return nil }
func (m *resizeTestHCSProcess) Close() error { return nil }

func TestResizeNarrowsConsoleDimensions(t *testing.T) {
	tests := []struct {
		name       string
		width      uint32
		height     uint32
		wantWidth  uint16
		wantHeight uint16
	}{
		{"normal_dimensions", 80, 24, 80, 24},
		{"maximum_uint16_width", 65535, 100, 65535, 100},
		{"width_wraps_to_zero", 65536, 100, 0, 100},
		{"width_wraps_to_one", 65537, 100, 1, 100},
		{"width_wraps_to_4464", 70000, 100, 4464, 100},
		{"height_wraps_to_zero", 80, 65536, 80, 0},
		{"both_dimensions_wrap_to_zero", 65536, 65536, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &resizeTestHCSProcess{}

			p := &process{
				id: "resize-test",
				ctr: &container{
					id: "container-test",
					client: &client{
						logger: log.G(context.Background()),
					},
				},
				hcsProcess: mock,
			}

			err := p.Resize(context.Background(), tc.width, tc.height)
			if err != nil {
				t.Fatalf("Resize returned unexpected error: %v", err)
			}

			t.Logf(
				"Input: width=%d, height=%d | Received: width=%d, height=%d | Expected: width=%d, height=%d",
				tc.width, tc.height,
				mock.width, mock.height,
				tc.wantWidth, tc.wantHeight,
			)

			if mock.width != tc.wantWidth {
				t.Errorf("width: got %d, want %d", mock.width, tc.wantWidth)
			}

			if mock.height != tc.wantHeight {
				t.Errorf("height: got %d, want %d", mock.height, tc.wantHeight)
			}
		})
	}
}
