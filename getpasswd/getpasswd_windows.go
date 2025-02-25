// +build windows

/*
 Copyright 2020-2025 Entrust Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package getpasswd

import (
	"bufio"
	"os"
	"strings"
	"syscall"
)

// refer SetConsoleMode documentation for the bit flags
// https://docs.microsoft.com/en-us/windows/console/setconsolemode
const ENABLE_ECHO_INPUT = 0x0004

// ReadPassword reads password from stdin
func ReadPassword() string {

	hStdIn := syscall.Handle(os.Stdin.Fd())
	var orgConsoleMode uint32

	err := syscall.GetConsoleMode(hStdIn, &orgConsoleMode)
	if err != nil {
		return ""
	}

	var newConsoleMode uint32 = (orgConsoleMode &^ ENABLE_ECHO_INPUT)

	err = setConsoleMode(hStdIn, newConsoleMode)
	defer setConsoleMode(hStdIn, orgConsoleMode)

	if err != nil {
		return ""
	}

	reader := bufio.NewReader(os.Stdin)
	passwd, _ := reader.ReadString('\n')
	if strings.HasSuffix(passwd, "\n") {
		passwd = passwd[:len(passwd)-1]
	}
	// check for \r just in case
	if strings.HasSuffix(passwd, "\r") {
		passwd = passwd[:len(passwd)-1]
	}

	return strings.TrimSpace(passwd)
}

func setConsoleMode(hConsole syscall.Handle, mode uint32) error {
	kernel32Dll := syscall.MustLoadDLL("kernel32.dll")
	procSetConsoleMode := kernel32Dll.MustFindProc("SetConsoleMode")
	ret, _, err := procSetConsoleMode.Call(uintptr(hConsole), uintptr(mode))
	if ret == 0 {
		return err
	}
	return nil
}
