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

package utils

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/facebook/openbmc/tools/flashy/lib/fileutils"
	"github.com/pkg/errors"
)

const ProcMtdFilePath = "/proc/mtd"

// memory information in bytes
type MemInfo struct {
	MemTotal uint64
	MemFree  uint64
}

// get memInfo
// note that this assumes kB units for MemFree and MemTotal
// it will fail otherwise
var GetMemInfo = func() (*MemInfo, error) {
	buf, err := fileutils.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, errors.Errorf("Unable to open /proc/meminfo: %v", err)
	}

	memInfoStr := string(buf)

	memFreeRegex := regexp.MustCompile(`(?m)^MemFree: +([0-9]+) kB$`)
	memTotalRegex := regexp.MustCompile(`(?m)^MemTotal: +([0-9]+) kB$`)

	memFreeMatch := memFreeRegex.FindStringSubmatch(memInfoStr)
	if len(memFreeMatch) < 2 {
		return nil, errors.Errorf("Unable to get MemFree in /proc/meminfo")
	}
	memFree, _ := strconv.ParseUint(memFreeMatch[1], 10, 64)
	memFree *= 1024 // convert to B

	memTotalMatch := memTotalRegex.FindStringSubmatch(memInfoStr)
	if len(memTotalMatch) < 2 {
		return nil, errors.Errorf("Unable to get MemTotal in /proc/meminfo")
	}
	memTotal, _ := strconv.ParseUint(memTotalMatch[1], 10, 64)
	memTotal *= 1024 // convert to B

	return &MemInfo{
		memTotal,
		memFree,
	}, nil
}

// function to aid logging and saving live stdout and stderr output
// from running command
// note that sequential execution is not guaranteed - race conditions
// might still exist
func logScanner(s *bufio.Scanner, ch chan struct{}, pre string, str *string) {
	for s.Scan() {
		t := fmt.Sprintf("%v\n", s.Text())
		log.Printf("%v%v", pre, t)
		*str = *str + t
	}
	close(ch)
}

// runs command and pipes live output
// returns exitcode, error, stdout (string), stderr (string) if non-zero/error returned or timed out
// returns 0, nil, stdout (string), stderr (string) if successfully run
var RunCommand = func(cmdArr []string, timeoutInSeconds int) (int, error, string, string) {
	start := time.Now()
	timeout := time.Duration(timeoutInSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCmdStr := strings.Join(cmdArr[:], " ")
	log.Printf("Running command '%v' with %v timeout", fullCmdStr, timeout)
	cmd := exec.CommandContext(ctx, cmdArr[0], cmdArr[1:]...)

	var stdoutStr, stderrStr string
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})
	go logScanner(stdoutScanner, stdoutDone, "stdout: ", &stdoutStr)
	go logScanner(stderrScanner, stderrDone, "stderr: ", &stderrStr)

	exitCode := 1 // show 1 if failed by default

	if err := cmd.Start(); err != nil {
		log.Printf("Command '%v' failed to start: %v", fullCmdStr, err)
		// failed to start, exit now
		return exitCode, err, stdoutStr, stderrStr
	}

	<-stdoutDone
	<-stderrDone

	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// The program exited with exit code != 0
			waitStatus := exitErr.Sys().(syscall.WaitStatus)
			exitCode = waitStatus.ExitStatus()
		} else {
			log.Printf("Could not get exit code, defaulting to '1'")
		}
	} else {
		// exit code should be 0
		waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = waitStatus.ExitStatus()
	}

	elapsed := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Command '%v' timed out after %v", fullCmdStr, timeout)
		// replace the err with the deadline exceeded error
		// instead of just signal: killed
		err = ctx.Err()
	} else {
		log.Printf("Command '%v' exited with code %v after %v", fullCmdStr, exitCode, elapsed)
	}

	return exitCode, err, stdoutStr, stderrStr
}

// calls RunCommand repeatedly until succeeded or maxAttempts is reached
// between attempts, an interval is applied
// returns the results from the first succeeding run or last tried run
var RunCommandWithRetries = func(cmdArr []string, timeoutInSeconds int, maxAttempts int, intervalInSeconds int) (int, error, string, string) {
	exitCode, err, stdoutStr, stderrStr := 1, errors.Errorf("Command failed to run"), "", ""

	if maxAttempts < 1 {
		err = errors.Errorf("Command failed to run: maxAttempts must be > 0 (got %v)", maxAttempts)
		return exitCode, err, stdoutStr, stderrStr
	}

	fullCmdStr := strings.Join(cmdArr[:], " ")
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempt %v of %v: Running command '%v' with timeout %vs and retry interval %vs",
			attempt, maxAttempts, fullCmdStr, timeoutInSeconds, intervalInSeconds)

		exitCode, err, stdoutStr, stderrStr = RunCommand(cmdArr, timeoutInSeconds)
		if err == nil {
			log.Printf("Attempt %v of %v succeeded", attempt, maxAttempts)
			break
		} else {
			log.Printf("Attempt %v of %v failed", attempt, maxAttempts)
			if attempt < maxAttempts {
				log.Printf("Sleeping for %vs before retrying", intervalInSeconds)
				sleepFunc(time.Duration(intervalInSeconds) * time.Second)
			} else {
				log.Printf("Max attempts (%v) reached. Returning with error.", maxAttempts)
			}
		}
	}
	return exitCode, err, stdoutStr, stderrStr
}

// check whether systemd is available
var SystemdAvailable = func() (bool, error) {
	const cmdlinePath = "/proc/1/cmdline"

	if fileutils.FileExists(cmdlinePath) {
		buf, err := fileutils.ReadFile(cmdlinePath)
		if err != nil {
			return false, errors.Errorf("%v exists but cannot be read: %v", cmdlinePath, err)
		}

		contains := strings.Contains(string(buf), "systemd")

		return contains, nil
	}

	return false, nil
}

// get OpenBMC version from /etc/issue
// examples: fbtp-v2020.09.1, wedge100-v2020.07.1
// WARNING: There is no guarantee that /etc/issue is well-formed
// in old images
var GetOpenBMCVersionFromIssueFile = func() (string, error) {
	const etcIssueVersionRegEx = `^OpenBMC Release (?P<version>[^\s]+)`

	etcIssueBuf, err := fileutils.ReadFile("/etc/issue")
	if err != nil {
		return "", errors.Errorf("Error reading /etc/issue: %v", err)
	}

	etcIssueMap, err := GetRegexSubexpMap(
		etcIssueVersionRegEx, string(etcIssueBuf))

	if err != nil {
		// does not match regex
		return "",
			errors.Errorf("Unable to get version from /etc/issue: %v", err)
	}

	version := etcIssueMap["version"]
	return version, nil
}
