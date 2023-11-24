package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/calvine/filejitsu/util/mock"
)

func TestRoundTripTar(t *testing.T) {
	type testCase struct {
		Name         string
		GetTarArgs   func(testRootDir, tarPath string) []string
		GetUntarArgs func(tarPath, untarPath string) []string
	}
	testCases := []testCase{
		{
			Name: "normal",
			GetTarArgs: func(testRootDir, tarPath string) []string {
				return []string{
					"tar",
					"-o",
					tarPath,
					testRootDir,
				}
			},
			GetUntarArgs: func(tarPath, untarPath string) []string {
				return []string{
					"tar",
					"-i",
					tarPath,
					"-u",
					untarPath,
				}
			},
		},
		{
			Name: "gzipped",
			GetTarArgs: func(testRootDir, tarPath string) []string {
				return []string{
					"tar",
					"-z",
					"-o",
					tarPath,
					testRootDir,
				}
			},
			GetUntarArgs: func(tarPath, untarPath string) []string {
				return []string{
					"tar",
					"-z",
					"-i",
					tarPath,
					"-u",
					untarPath,
				}
			},
		},
		{
			Name: "encrypted",
			GetTarArgs: func(testRootDir, tarPath string) []string {
				return []string{
					"tar",
					"-e",
					"-s",
					"test1",
					"-o",
					tarPath,
					testRootDir,
				}
			},
			GetUntarArgs: func(tarPath, untarPath string) []string {
				return []string{
					"tar",
					"-e",
					"-s",
					"test1",
					"-i",
					tarPath,
					"-u",
					untarPath,
				}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			testRootDir, content, cleanup, err := mock.MockDirTree("fjtest_files")
			if err != nil {
				t.Errorf("failed to create test dir tree: %v", err)
			}
			defer cleanup()
			tmpDir := os.TempDir()
			tarPath := filepath.Join(tmpDir, "output.tar")
			if err != nil {
				t.Errorf("failed to create mock dir tree: %v", err)
				return
			}
			tarCmd := SetupCommand("", "")
			tarCmd.SetArgs(tc.GetTarArgs(testRootDir, tarPath))
			if err = tarCmd.Execute(); err != nil {
				t.Errorf("failed to run tar on dir: %v", err)
				return
			}
			untarPath := filepath.Join(tmpDir, "test_untar")
			untarCmd := SetupCommand("", "")
			untarCmd.SetArgs(tc.GetUntarArgs(tarPath, untarPath))
			defer func() {
				os.Remove(tarPath)
				os.RemoveAll(untarPath)
			}()
			if err := untarCmd.Execute(); err != nil {
				t.Errorf("failed to run untar on tar file: %v", err)
				return
			}
			// compare to content...
			err = mock.ConfirmContentMapMatches(untarPath, content)
			if err != nil {
				t.Errorf("failed in comparison of untar'ed files: %v", err)
			}
		})
	}
}
