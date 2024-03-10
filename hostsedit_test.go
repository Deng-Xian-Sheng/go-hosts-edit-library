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
	"os"
	"testing"
)

// 创建测试用的hosts文件
func createTestHostsFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("./", "_test_hosts")
	if err != nil {
		return "", err
	}
	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", err
	}
	err = tmpFile.Close()
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

// 测试New函数
func TestNew(t *testing.T) {
	hostsContent := `
127.0.0.1 localhost
::1 localhost
# Comment line
`
	filePath, err := createTestHostsFile(hostsContent)
	if err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}
	defer os.Remove(filePath)

	hostsEdit, err := New(filePath, false)
	if err != nil {
		t.Errorf("New() error = %v, wantErr = false", err)
	}

	if len(hostsEdit.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(hostsEdit.Lines))
	}
}

// 测试Get方法
func TestGet(t *testing.T) {
	hostsContent := `
127.0.0.1 localhost
::1 ipv6host
# Comment line
`
	filePath, err := createTestHostsFile(hostsContent)
	if err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}
	defer os.Remove(filePath)

	hostsEdit, _ := New(filePath, false)

	ip, exists := hostsEdit.Get("localhost")
	if !exists || ip != "127.0.0.1" {
		t.Errorf("Get(localhost) = %v, %v; want %v, %v", ip, exists, "127.0.0.1", true)
	}

	ip, exists = hostsEdit.Get("ipv6host")
	if !exists || ip != "::1" {
		t.Errorf("Get(ipv6host) = %v, %v; want %v, %v", ip, exists, "::1", true)
	}
}

// 测试Exists方法
func TestExists(t *testing.T) {
	hostsContent := `
127.0.0.1 localhost
::1 ipv6host
# Comment line
`
	filePath, err := createTestHostsFile(hostsContent)
	if err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}
	defer os.Remove(filePath)

	hostsEdit, _ := New(filePath, false)

	if !hostsEdit.Exists("localhost") {
		t.Errorf("Exists(localhost) = false; want true")
	}

	if !hostsEdit.Exists("ipv6host") {
		t.Errorf("Exists(ipv6host) = false; want true")
	}

	if hostsEdit.Exists("nonexistent") {
		t.Errorf("Exists(nonexistent) = true; want false")
	}
}

// 测试Edit方法
func TestEdit(t *testing.T) {
	hostsContent := `
127.0.0.1 localhost
::1 ipv6host
# Comment line
`
	filePath, err := createTestHostsFile(hostsContent)
	if err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}
	defer os.Remove(filePath)

	hostsEdit, _ := New(filePath, false)

	// 添加新主机记录
	err = hostsEdit.Edit("newhost", "127.0.0.2")
	if err != nil {
		t.Errorf("Edit(newhost, 127.0.0.2) failed with error: %v", err)
	}

	// 更新现有主机记录的IP地址
	err = hostsEdit.Edit("localhost", "127.0.0.3")
	if err != nil {
		t.Errorf("Edit(localhost, 127.0.0.3) failed with error: %v", err)
	}

	// 重新加载hosts文件以验证更改
	updatedHostsEdit, _ := New(filePath, false)
	if !updatedHostsEdit.Exists("newhost") {
		t.Errorf("Edit failed to add newhost")
	}

	newIP, exists := updatedHostsEdit.Get("localhost")
	if !exists || newIP != "127.0.0.3" {
		t.Errorf("Edit failed to update localhost, got IP %v", newIP)
	}
}
