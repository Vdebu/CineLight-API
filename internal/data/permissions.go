package data

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// 使用自定义类型存储从数据库中查询到的权限数据
type Permissions []string

// 检测目前要求的权限是否在查询到的权限中
func (p Permissions) Include(code string) bool {
	for i := range p {
		log.Println(i)
		if p[i] == code {
			return true
		}
	}
	return false
}

// 解耦数据库连接池
type PermissionModel struct {
	db *sql.DB
}

// 针对一个用户查找其所拥有的权限
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	stmt := `
			SELECT permissions.code
			FROM permissions
			INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
			INNER JOIN users ON users_permissions.user_id = users.id
			WHERE users.id = $1`
	//创建五秒查操作超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 回收资源
	defer cancel()
	// 尝试查询数据
	// 使用QueryContext查询多条记录 返回rows用于后续遍历提取
	rows, err := m.db.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}
	// 读取完毕后关闭资源(忘記寫defer過早關閉的了數據集導致讀取不到數據後續權限驗證失敗)
	defer rows.Close()
	// 创建自定义容器存储提取到的数据
	var permissions Permissions
	for rows.Next() {
		// 創建容器存儲查詢到的數據
		var permission string
		// 將數據提取到容器中
		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		log.Println(permission)
		// 将读取到的记录加入切片
		permissions = append(permissions, permission)
	}
	// 检查读取数据的过程中是否发生过错误
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// 没发生错误则返回存储结果
	return permissions, nil
}
