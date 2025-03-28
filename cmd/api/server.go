package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 初始化并启动服务器的模块
func (app *application) server() error {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port), // 使用字符串初始化端口(:%d)
		Handler:      app.routers(),                       // 初始化路由
		IdleTimeout:  time.Minute,                         // 初始化各种操作的超时时间
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		//ErrorLog:     log.New(logger, "", 0), // 由于实现了io.Writer接口可以直接用自定义logger创建新logger
	}
	// 创建error通道监听graceful Shutdown返回的错误信息
	shutdownError := make(chan error)
	// 启动goroutine监听服务器相关的信号
	go func() {
		// 带缓冲的通道用于接受信号 避免错过终止信号
		quit := make(chan os.Signal, 1)
		// 使用signal.Notify监听中断与结束信号
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// 从通道中读取信号 初始情况下是阻塞的
		s := <-quit
		// 以JSON的形式输出捕获到的信号
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})
		// 创建5秒超时的ctx用于后续服务器资源的关闭(为已经传入的请求创造5秒的完成时间)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		//确保ctx资源回收
		defer cancel()
		// 尝试关闭判断是否发生错误
		err := srv.Shutdown(ctx)
		if err != nil {
			// 若发生错误直接将其存入通道并返回
			shutdownError <- err
			return
		}
		// 服务器关闭成功等待后台进程全部结束完毕
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})
		// 使用WaitGroup等待进行完成
		app.wg.Wait()
		// 将nil存入ShutdownErr
		shutdownError <- nil
	}()
	app.logger.PrintInfo("starting the server", map[string]string{
		"addr": srv.Addr,       // 输出端口信息
		"env":  app.config.env, // 输出开发环境信息
	})
	// 启动服务器 启动服务器后会直接进入阻塞 直接收到让服务器停止的信号
	err := srv.ListenAndServe()
	// 在进入优雅退出后ListenAndServe会接收到http.ErrServerClosed错误
	// 检查错误类型决定是否返回
	if !errors.Is(err, http.ErrServerClosed) {
		// 若不是优雅退出造成的错误则返回处理
		return err
	}
	// 监听优雅退出是否开始
	err = <-shutdownError
	if err != nil {
		// 退出过程发生错误进行返回
		return err
	}
	// 到这里就已经实现优雅退出了 输出相关的成功信息
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
