# go-hosts-edit-library
host file edit library by Golang. Golang的hosts文件编辑库。

## Quick start

```go
package main

import hostsedit "github.com/Deng-Xian-Sheng/go-hosts-edit-library"

func main() {
	// Create a file to ./hosts , as the following:
	/*
	   127.0.0.1 localhost
	   ::1 localhost
	   # Comment line
	   1.1.1.1 google.com
	   2.2.2.2 baidu.com
	*/

	// init
	// 初始化
	hostEdit,err :=  hostsedit.New("./hosts",false)
	if err != nil {
		panic(err)
	}

	// get ip by host
	// 通过主机获取ip
	v,ok := hostEdit.Get("google.com")
	if ok {
		println(v)
	}

	// edit ip by host, If the host is not in the hosts file, it will create a new one
	// 通过主机编辑ip，如果主机不在hosts文件中，它会新建一条
	err = hostEdit.Edit("google.com","3.3.3.3")
	if err != nil {
		panic(err)
	}

	ok = hostEdit.Exists("baidu.com")
	if ok {
		println("baidu.com exists hosts file")
	}else{
		println("baidu.com not exists hosts file")
	}

    // not exists no error
	err = hostEdit.Delete("baidu.com")
	if err != nil {
		panic(err)
	}
}

```