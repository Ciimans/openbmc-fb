/**
 * Copyright 2020-present Facebook. All Rights Reserved.
 *
 * This program file is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the
 * Free Software Foundation; version 2 of the License.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
 * for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program in a file named COPYING; if not, write to the
 * Free Software Foundation, Inc.,
 * 51 Franklin Street, Fifth Floor,
 * Boston, MA 02110-1301 USA
 */

package flashutils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/facebook/openbmc/tools/flashy/lib/flash/flashutils/devices"
	"github.com/facebook/openbmc/tools/flashy/lib/utils"
	"github.com/facebook/openbmc/tools/flashy/tests"
	"github.com/pkg/errors"
)

type partialRunCommandReturn struct {
	err    error
	stdout string
}

func TestGetVbootUtilContents(t *testing.T) {
	// mock and defer restore RemoveFile, RunCommand, IsVbootSystem and
	// IsELFFile
	removeFileOrig := utils.RemoveFile
	runCommandOrig := utils.RunCommand
	isVbootSystemOrig := isVbootSystem
	isELFFileOrig := utils.IsELFFile
	defer func() {
		utils.RemoveFile = removeFileOrig
		utils.RunCommand = runCommandOrig
		isVbootSystem = isVbootSystemOrig
		utils.IsELFFile = isELFFileOrig
	}()

	// removeFile is mocked to return nil
	utils.RemoveFile = func(_ string) error {
		return nil
	}

	cases := []struct {
		name          string
		isVbootSystem bool
		isELF         bool
		runCmdRet     partialRunCommandReturn
		wantCmd       string
		want          string
		wantErr       error
	}{
		{
			name:          "not a vboot system",
			isVbootSystem: false,
			isELF:         true,
			runCmdRet:     partialRunCommandReturn{},
			wantCmd:       "",
			want:          "",
			wantErr:       errors.Errorf("Not a vboot system"),
		},
		{
			name:          "is ELF file",
			isVbootSystem: true,
			isELF:         true,
			runCmdRet: partialRunCommandReturn{
				nil,
				"foobar",
			},
			wantCmd: vbootUtilPath,
			want:    "foobar",
			wantErr: nil,
		},
		{
			name:          "not ELF",
			isVbootSystem: true,
			isELF:         false,
			runCmdRet: partialRunCommandReturn{
				nil,
				"foobar",
			},
			wantCmd: fmt.Sprintf("bash %v", vbootUtilPath),
			want:    "foobar",
			wantErr: nil,
		},
		{
			name:          "failed",
			isVbootSystem: true,
			isELF:         true,
			runCmdRet: partialRunCommandReturn{
				errors.Errorf("cmd err"),
				"",
			},
			wantCmd: vbootUtilPath,
			want:    "",
			wantErr: errors.Errorf("Unable to get vboot-util info: cmd err"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotCmd := ""
			isVbootSystem = func() bool {
				return tc.isVbootSystem
			}
			utils.RunCommand = func(cmdArr []string, timeoutInSeconds int) (int, error, string, string) {
				gotCmd = strings.Join(cmdArr, " ")
				retErr := tc.runCmdRet.err
				stdout := tc.runCmdRet.stdout
				return 0, retErr, stdout, ""
			}
			utils.IsELFFile = func(filename string) bool {
				if filename != vbootUtilPath {
					t.Errorf("want vboot util path '%v' got '%v'", vbootUtilPath, filename)
				}
				return tc.isELF
			}
			got, err := getVbootUtilContents()

			if tc.want != got {
				t.Errorf("want '%v' got '%v'", tc.want, got)
			}
			tests.CompareTestErrors(tc.wantErr, err, t)
			if tc.wantCmd != gotCmd {
				t.Errorf("want cmd '%v' got '%v'", tc.wantCmd, gotCmd)
			}
		})
	}
}

func TestGetVbootEnforcement(t *testing.T) {
	// mock and defer restore IsVbootSystem, ReadFile and getVbootUtilContents
	isVbootSystemOrig := isVbootSystem
	readFileOrig := utils.ReadFile
	getVbootUtilContentsOrig := getVbootUtilContents
	defer func() {
		isVbootSystem = isVbootSystemOrig
		utils.ReadFile = readFileOrig
		getVbootUtilContents = getVbootUtilContentsOrig
	}()

	cases := []struct {
		name              string
		isVbootSystem     bool
		procMtdContents   string
		procMtdReadErr    error
		vbootUtilContents string
		vbootUtilGetErr   error
		want              vbootEnforcementEnum
		wantErr           error
	}{
		{
			name:              "Not a vboot system",
			isVbootSystem:     false,
			procMtdContents:   "",
			procMtdReadErr:    nil,
			vbootUtilContents: "",
			vbootUtilGetErr:   nil,
			want:              vbootNone,
			wantErr:           nil,
		},
		{
			name:              "example wedge100 /proc/mtd, no romx, mock vboot system",
			isVbootSystem:     true,
			procMtdContents:   tests.ExampleWedge100ProcMtdFile,
			procMtdReadErr:    nil,
			vbootUtilContents: "",
			vbootUtilGetErr:   nil,
			want:              vbootNone,
			wantErr:           nil,
		},
		{
			name:              "tiogapass1 example",
			isVbootSystem:     true,
			procMtdContents:   tests.ExampleTiogapass1ProcMtdFile,
			procMtdReadErr:    nil,
			vbootUtilContents: tests.ExampleTiogapass1VbootUtilFile,
			vbootUtilGetErr:   nil,
			want:              vbootHardwareEnforce,
			wantErr:           nil,
		},
		{
			name:              "/proc/mtd read err",
			isVbootSystem:     true,
			procMtdContents:   "",
			procMtdReadErr:    errors.Errorf("proc mtd read err"),
			vbootUtilContents: "",
			vbootUtilGetErr:   nil,
			want:              vbootNone,
			wantErr:           errors.Errorf("Unable to read '/proc/mtd': proc mtd read err"),
		},
		{
			name:              "getVbootUtilContents err",
			isVbootSystem:     true,
			procMtdContents:   "romx",
			procMtdReadErr:    nil,
			vbootUtilContents: "",
			vbootUtilGetErr:   errors.Errorf("getVbootUtilContents err"),
			want:              vbootNone,
			wantErr:           errors.Errorf("Unable to read vboot-util contents: getVbootUtilContents err"),
		},
		{
			name:            "software enforce example",
			isVbootSystem:   true,
			procMtdContents: "romx",
			procMtdReadErr:  nil,
			vbootUtilContents: `Flags hardware_enforce:  0x00
Flags software_enforce:  0x01`,
			vbootUtilGetErr: nil,
			want:            vbootSoftwareEnforce,
			wantErr:         nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isVbootSystem = func() bool {
				return tc.isVbootSystem
			}
			utils.ReadFile = func(filename string) ([]byte, error) {
				if filename != utils.ProcMtdFilePath {
					t.Errorf("filename: want '%v' got '%v'",
						utils.ProcMtdFilePath, filename)
				}
				return []byte(tc.procMtdContents), tc.procMtdReadErr
			}
			getVbootUtilContents = func() (string, error) {
				return tc.vbootUtilContents, tc.vbootUtilGetErr
			}
			got, err := getVbootEnforcement()
			if tc.want != got {
				t.Errorf("want '%v' got '%v'", tc.want, got)
			}
			tests.CompareTestErrors(tc.wantErr, err, t)
		})
	}
}

func TestPatchImageWithLocalBootloader(t *testing.T) {
	// save log output into buf for testing
	var buf bytes.Buffer
	log.SetOutput(&buf)
	// mock and defer restore ReplaceFirstNBytes
	replaceFirstNBytesOrig := utils.ReplaceFirstNBytes
	defer func() {
		log.SetOutput(os.Stderr)
		utils.ReplaceFirstNBytes = replaceFirstNBytesOrig
	}()
	const imageFilePath = "imageFilePath"
	const flashDevicePath = "flashDevicePath"
	const testOffset = 42

	cases := []struct {
		name             string
		replaceNBytesErr error
		logContainsSeq   []string
		want             error
	}{
		{
			name:             "patch success",
			replaceNBytesErr: nil,
			logContainsSeq: []string{
				"===== WARNING: PATCHING IMAGE FILE =====",
				"This vboot system has 42B RO offset in mtd, patching image file with offset.",
				"Successfully patched image file",
			},
			want: nil,
		},
		{
			name:             "patch failed",
			replaceNBytesErr: errors.Errorf("Unable to replace bytes"),
			logContainsSeq:   []string{},
			want:             errors.Errorf("Unable to patch image: Unable to replace bytes"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf = bytes.Buffer{}
			utils.ReplaceFirstNBytes = func(dstPath, srcPath string, n int) error {
				if dstPath != imageFilePath {
					t.Errorf("dstPath: want '%v' got '%v'", imageFilePath, dstPath)
				}
				if srcPath != flashDevicePath {
					t.Errorf("srcPath: want '%v' got '%v'", flashDevicePath, srcPath)
				}
				if n != testOffset {
					t.Errorf("n: want '%v' got '%v'", testOffset, n)
				}
				return tc.replaceNBytesErr
			}
			got := patchImageWithLocalBootloader(imageFilePath, flashDevicePath, testOffset)

			tests.CompareTestErrors(tc.want, got, t)
			tests.LogContainsSeqTest(buf.String(), tc.logContainsSeq, t)
		})

	}

}

func TestIsVbootImagePatchingRequired(t *testing.T) {
	// mock and defer restore getVbootReinforcement
	getVbootEnforcementOrig := getVbootEnforcement
	defer func() {
		getVbootEnforcement = getVbootEnforcementOrig
	}()
	cases := []struct {
		name                 string
		vbootEnforcement     vbootEnforcementEnum
		vbootEnforcementErr  error
		flashDeviceSpecifier string
		want                 bool
		wantErr              error
	}{
		{
			name:                 "No vboot enforcement",
			vbootEnforcement:     vbootNone,
			vbootEnforcementErr:  nil,
			flashDeviceSpecifier: "flash1",
			want:                 false,
			wantErr:              nil,
		},
		{
			name:                 "software vboot enforcement",
			vbootEnforcement:     vbootSoftwareEnforce,
			vbootEnforcementErr:  nil,
			flashDeviceSpecifier: "flash1",
			want:                 false,
			wantErr:              nil,
		},
		{
			name:                 "hardware vboot enforcement, flash1",
			vbootEnforcement:     vbootHardwareEnforce,
			vbootEnforcementErr:  nil,
			flashDeviceSpecifier: "flash1",
			want:                 true,
			wantErr:              nil,
		},
		{
			name:                 "hardware vboot enforcement, flash0 (not required)",
			vbootEnforcement:     vbootHardwareEnforce,
			vbootEnforcementErr:  nil,
			flashDeviceSpecifier: "flash0",
			want:                 false,
			wantErr:              nil,
		},
		{
			name:                 "error getting vboot enforcement",
			vbootEnforcement:     vbootNone,
			vbootEnforcementErr:  errors.Errorf("vboot enforcement err"),
			flashDeviceSpecifier: "flash0",
			want:                 false,
			wantErr:              errors.Errorf("Unable to get vboot enforcement: vboot enforcement err"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getVbootEnforcement = func() (vbootEnforcementEnum, error) {
				return tc.vbootEnforcement, tc.vbootEnforcementErr
			}
			got, err := isVbootImagePatchingRequired(tc.flashDeviceSpecifier)
			if tc.want != got {
				t.Errorf("want %v got %v", tc.want, got)
			}
			tests.CompareTestErrors(tc.wantErr, err, t)
		})
	}
}

func TestVbootPatchImageBootloaderIfNeeded(t *testing.T) {
	// mock and defer restore isVbootSystem, isVbootImagePatchingRequired
	// and patchImageWithLocalBootloader
	isVbootSystemOrig := isVbootSystem
	isVbootImagePatchingRequiredOrig := isVbootImagePatchingRequired
	patchImageWithLocalBootloaderOrig := patchImageWithLocalBootloader
	defer func() {
		isVbootSystem = isVbootSystemOrig
		isVbootImagePatchingRequired = isVbootImagePatchingRequiredOrig
		patchImageWithLocalBootloader = patchImageWithLocalBootloaderOrig
	}()

	exampleFlashDevice := &devices.FlashDevice{
		"mtd",
		"flash1",
		"/dev/mtd5",
		uint64(42),
	}

	cases := []struct {
		name             string
		isVboot          bool
		patchRequired    bool
		patchRequiredErr error
		patchErr         error
		want             error
	}{
		{
			name:             "vboot patch required & succeeded",
			isVboot:          true,
			patchRequired:    true,
			patchRequiredErr: nil,
			patchErr:         nil,
			want:             nil,
		},
		{
			name:             "vboot patch not required",
			isVboot:          true,
			patchRequired:    false,
			patchRequiredErr: nil,
			patchErr:         nil,
			want:             nil,
		},
		{
			name:             "not a vboot system",
			isVboot:          false,
			patchRequired:    false,
			patchRequiredErr: nil,
			patchErr:         nil,
			want:             errors.Errorf("Not a vboot system, cannot run vboot remediation"),
		},
		{
			name:             "patch required check failed",
			isVboot:          true,
			patchRequired:    false,
			patchRequiredErr: errors.Errorf("check fail"),
			patchErr:         nil,
			want:             errors.Errorf("Unable to determine whether image patching is required: check fail"),
		},
		{
			name:             "patch required and failed",
			isVboot:          true,
			patchRequired:    true,
			patchRequiredErr: nil,
			patchErr:         errors.Errorf("patch fail"),
			want:             errors.Errorf("Failed to patch image with local bootloader: patch fail"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isVbootSystem = func() bool {
				return tc.isVboot
			}
			isVbootImagePatchingRequired = func(flashDeviceSpecifier string) (bool, error) {
				if flashDeviceSpecifier != exampleFlashDevice.Specifier {
					t.Errorf("flash device specifier: want '%v' got '%v'",
						exampleFlashDevice.Specifier, flashDeviceSpecifier)
				}
				return tc.patchRequired, tc.patchRequiredErr
			}
			patchImageWithLocalBootloader = func(imageFilePath, flashDeviceFilePath string, offsetBytes int) error {
				if imageFilePath != "x" {
					t.Errorf("image file path: want 'x' got '%v'", imageFilePath)
				}
				if flashDeviceFilePath != "/dev/mtd5" {
					t.Errorf("image file path: want '/dev/mtd5' got '%v'", flashDeviceFilePath)
				}
				if offsetBytes != vbootOffset {
					t.Errorf("offsetBytes: want '%v' got '%v'", vbootOffset, offsetBytes)
				}
				return tc.patchErr
			}
			got := VbootPatchImageBootloaderIfNeeded("x", exampleFlashDevice)
			tests.CompareTestErrors(tc.want, got, t)
		})
	}
}
