package main

import (
	"fmt"
	"net/http"
)

// 返回服务器的状态 运行环境 版本
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status:available")
	fmt.Fprintf(w, "environment:%s\n", app.config.env)
	fmt.Fprintf(w, "version:%s\n", version)
}
