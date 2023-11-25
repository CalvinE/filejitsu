package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/calvine/filejitsu/util/mock"
)

func TestRoundTripTar(t *testing.T) {
	type testCase struct {
		Name                 string
		AdditionalInputPaths []string
		GetTarArgs           func(inputPaths []string, tarPath string) []string
		GetUntarArgs         func(tarPath, untarPath string) []string
	}
	testCases := []testCase{
		{
			Name: "normal",
			GetTarArgs: func(inputPaths []string, tarPath string) []string {
				args := []string{
					"tar",
					"-o",
					tarPath,
				}
				args = append(args, inputPaths...)
				return args
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
			GetTarArgs: func(inputPaths []string, tarPath string) []string {
				args := []string{
					"tar",
					"-z",
					"-o",
					tarPath,
				}
				args = append(args, inputPaths...)
				return args
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
			GetTarArgs: func(inputPaths []string, tarPath string) []string {
				args := []string{
					"tar",
					"-e",
					"-p",
					"test1",
					"-o",
					tarPath,
				}
				args = append(args, inputPaths...)
				return args
			},
			GetUntarArgs: func(tarPath, untarPath string) []string {
				return []string{
					"tar",
					"-e",
					"-p",
					"test1",
					"-i",
					tarPath,
					"-u",
					untarPath,
				}
			},
		},
		// {
		// 	Name: "encrypted dir and single file",
		// 	AdditionalInputPaths: []string{
		// 		"/home/calvin/code/filejitsu/README.md",
		// 	},
		// 	GetTarArgs: func(inputPaths []string, tarPath string) []string {
		// 		args := []string{
		// 			"tar",
		// 			"-e",
		// 			"-p",
		// 			"test1",
		// 			"-o",
		// 			tarPath,
		// 		}
		// 		args = append(args, inputPaths...)
		// 		return args
		// 	},
		// 	GetUntarArgs: func(tarPath, untarPath string) []string {
		// 		return []string{
		// 			"tar",
		// 			"-e",
		// 			"-p",
		// 			"test1",
		// 			"-i",
		// 			tarPath,
		// 			"-u",
		// 			untarPath,
		// 		}
		// 	},
		// },
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
			tarCmd := SetupCommand("", "", "")
			inputPaths := []string{testRootDir}
			if len(tc.AdditionalInputPaths) > 0 {
				inputPaths = append(inputPaths, tc.AdditionalInputPaths...)
			}
			tarCmd.SetArgs(tc.GetTarArgs(inputPaths, tarPath))
			if err := tarCmd.Execute(); err != nil {
				t.Errorf("failed to run tar on dir: %v", err)
				return
			}
			untarPath := filepath.Join(tmpDir, "test_untar")
			untarCmd := SetupCommand("", "", "")
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

func TestTarHelp(t *testing.T) {
	tarCmd := SetupCommand("", "", "")
	tarCmd.SetArgs([]string{
		"tar",
		"-h",
	})
	if err := tarCmd.Execute(); err != nil {
		t.Errorf("failed to run tar on dir: %v", err)
		return
	}
}
