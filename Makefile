# 检测当前的操作系统根据操作系统选择后续执行的指令
ifeq ($(OS),Windows_NT)
    CONFIRM_CMD = choice /c yn /n /m "Are you sure? [y/N]"
    CONFIRM_CHECK = if errorlevel 2 exit 1
else
    CONFIRM_CMD = read -p "Are you sure? [y/N] " ans && [ "$${ans}" = "y" ]
endif
# 展示帮助信息
.PHONY: help:
	@echo 'Usage'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
# 询问用户是否确认执行
.PHONY: confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [$${ans:-N} = y]
# 启动API
.PHONY: run/api:
	go run ./cmd/api
# 连接数据库
.PHONY: db/psql:
	psql ${GREENLIGHT_DB_DSN}
# 创建新的迁移文件
.PHONY: /migration/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}
# 执行迁移指令
.PHONY: db/migrations/up:confirm
	@echo 'Running up migrations'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up