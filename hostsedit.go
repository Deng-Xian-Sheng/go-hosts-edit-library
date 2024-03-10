// host file edit library by Golang.
// Copyright (C) 2024 CanQi Jin

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.


package hostedit

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

/*
#
x.x.x.x xxx xxx xxx
x.x.x.x xxx
*/

type Line struct {
	IsComment           bool
	UndefinedRowsRawStr string
	IP                  string
	Host                map[string]struct{} // 注意，会使多个主机之间无序，但是这貌似是不可避免的。
}

// HostsEdit represents the entire hosts file and provides methods to manipulate it.
type HostsEdit struct {
	Lines    []*Line
	FilePath string
}

// New loads the hosts file from the specified path and returns a HostsEdit instance.
// isParse 是否进行严格的语法分析，如果启用则不容忍注释行以外的重复的主机、不规范的主机的条目，遇到此类会报错。但操作系统在这种情况下往往不会报错，与操作系统的行为不符。
func New(filePath string, isParse bool) (*HostsEdit, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []*Line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineText := scanner.Text()
		line := Line{
			Host: make(map[string]struct{}),
		}

		lineText = strings.TrimSpace(lineText)

		if lineText == "" {
			continue
		}

		if strings.HasPrefix(lineText, "#") {
			line.IsComment = true
			lineText = strings.TrimSpace(strings.TrimPrefix(lineText, "#"))
		}

		entries := strings.Fields(lineText)
		if len(entries) >= 2 && net.ParseIP(entries[0]) != nil {
			line.IP = entries[0]
			for _, v := range entries[1:] {
				line.Host[v] = struct{}{}
			}
		} else {
			line.UndefinedRowsRawStr = lineText
		}

		lines = append(lines, &line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if isParse {
		err = parse(lines)
		if err != nil {
			return nil, err
		}
	}

	return &HostsEdit{Lines: lines, FilePath: filePath}, nil
}

func parse(lines []*Line) (err error) {
	allHost := make(map[string]struct{})
	for _, line := range lines {
		if line.IsComment {
			continue
		}
		if line.UndefinedRowsRawStr != "" {
			return errors.New("not comment but UndefinedRowsRawStr")
		}
		for k := range line.Host {
			if _, ok := allHost[k]; !ok {
				allHost[k] = struct{}{}
			} else {
				return errors.New("host repeat")
			}
		}
	}
	return nil
}

// Get returns the IP address of the specified host.
func (h *HostsEdit) Get(host string) (string, bool) {
	for _, line := range h.Lines {
		if line.IsComment || line.UndefinedRowsRawStr != "" {
			continue
		}
		if _, exists := line.Host[host]; exists {
			return line.IP, true
		}
	}
	return "", false
}

// Exists checks if the specified host exists in the hosts file.
func (h *HostsEdit) Exists(host string) bool {
	_, exists := h.Get(host)
	return exists
}

// Edit adds or updates the specified host with the given IP address.
func (h *HostsEdit) Edit(host, ip string) (err error) {
	for _, line := range h.Lines {
		if line.IsComment || line.UndefinedRowsRawStr != "" {
			continue
		}
		if _, exists := line.Host[host]; exists {
			if line.IP == ip {
				return
			}
			if len(line.Host) > 1 {
				delete(line.Host, host)
			} else {
				line.IP = ip
				err = saveToFile(h.Lines, h.FilePath)
				if err != nil {
					return err
				}
				return
			}
		}
	}

	for _, line := range h.Lines {
		if line.IsComment || line.UndefinedRowsRawStr != "" {
			continue
		}
		if line.IP == ip {
			line.Host[host] = struct{}{}
			err = saveToFile(h.Lines, h.FilePath)
			if err != nil {
				return err
			}
			return
		}
	}

	// 头部追加，防止因主机重导致操作系统识别的时候忽视
	h.Lines = append([]*Line{{
		IsComment:           false,
		UndefinedRowsRawStr: "",
		IP:                  ip,
		Host: map[string]struct{}{
			host: {},
		},
	}}, h.Lines...)

	err = saveToFile(h.Lines, h.FilePath)
	if err != nil {
		return err
	}

	return
}

// saveToFile writes the current hosts file configuration back to disk.
func saveToFile(lines []*Line, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range lines {
		if line.IsComment {
			_, err := fmt.Fprint(file, "# ")
			if err != nil {
				return err
			}
		}
		if line.UndefinedRowsRawStr != "" {
			_, err := fmt.Fprint(file, line.UndefinedRowsRawStr)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprint(file, line.IP, " ")
			if err != nil {
				return err
			}

			count := 1
			for k := range line.Host {
				split := " "
				if count == len(line.Host) {
					split = ""
				}
				_, err := fmt.Fprint(file, k, split)
				if err != nil {
					return err
				}
				count++
			}
		}

		_, err := fmt.Fprint(file, "\n")
		if err != nil {
			return err
		}
	}

	return nil
}
