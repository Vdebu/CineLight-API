package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"greenlight.vdebu.net/internal/validator"
	"time"
)

// 使用结构体存储基本信息
type Movie struct {
	ID        int64     `json:"id"`                // 唯一标识
	CreatedAt time.Time `json:"-"`                 // 加入数据库的时间 对数据库中的创建时间字段进行隐藏
	Title     string    `json:"title"`             // 标题
	Year      int32     `json:"year,omitempty"`    // 发行时间
	Runtime   Runtime   `json:"runtime,omitempty"` // 时长 使用自定义类型存储播放时长(实现了json.Marshal接口)生成自定义的格式化信息
	Genres    []string  `json:"genres,omitempty"`  // 标签 对发行时间时长标签进行空值隐藏("",0,nil或空slice,map)
	Version   int32     `json:"version"`           // 版本信息从1开始 当电影信息更新版本信息会自动递增
}

// 用于检测电影结构体的各个字段是否有效
func ValidateMovie(v *validator.Validator, movie *Movie) {
	// 对输入的各个字段进行检查
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "year must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	// 查看是否有标签是重复的
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}

// 创建模型存储数据库连接池
type MovieModel struct {
	db *sql.DB
}

// 向数据库插入新数据
func (m *MovieModel) Insert(movie *Movie) error {
	// 插入的SQL语句
	// $N是Postgresql 特殊占位符
	stmt := `INSERT INTO movies (title,year,runtime,genres)
				VALUES ($1,$2,$3,$4)
				RETURNING id,created_at,version` // 提取数据库自动生成的信息写入传入的结构体
	// 插入三个以上的数据(三个以上的占位符) 使用[]interface{}进行存储并作为参数传入 注意进行类型转换
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	// 将RETURNING返回的数据插入传进来的数据(默认为空值)
	// 创建ctx实现DeadLine
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return m.db.QueryRowContext(ctx, stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// 使用id从数据库中查找数据
func (m *MovieModel) Get(id int64) (*Movie, error) {
	// 避免进行不必要的数据库查询
	if id < 1 {
		// 记录是从1开始的
		return nil, ErrRecordNotFound
	}
	stmt := `SELECT id,created_at,title,year,runtime,genres,version 
			FROM movies
			WHERE id = $1`
	// 存储查询到的数据
	var movie Movie
	// 使用ContextWithTimeout定义查询进行的最长时间
	// 基板context与duration
	// 这里定义完成就是已经开始计时了
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	// 在Get方法返回前调用cancel回收资源释放内存
	defer cancel()
	// 将ctx传入设置DeadLine
	// 使用pq.Array()对查询到的数据进行转换以后存入结构体 使用空的字节数组存储pq_sleep返回的数据
	err := m.db.QueryRowContext(ctx, stmt, id).Scan(
		&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		// 判断是不是sql的no row 错误
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		// 返回其他错误
		return nil, err
	}
	return &movie, nil
}

// 根据新数据更新数据库
func (m *MovieModel) Update(movie *Movie) error {
	// 通过id更新数据 更新成功后返回新的版本信息
	// 基于表中的VERSION字段实现乐观锁
	stmt := `UPDATE movies
			SET title = $1,year = $2,runtime = $3,genres = $4,version = version + 1
			WHERE id = $5 AND version = $6
			RETURNING version`
	// 存储要使用的参数
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}
	// 创建ctx实现DeadLine
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	// 判断当前目标的VERSION字段是否发生了变化
	err := m.db.QueryRowContext(ctx, stmt, args...).Scan(&movie.Version)
	if err != nil {
		// 判断错误类型 如果是NoRows则说明VERSION不一致 发生了冲突
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// 根据id从数据库中删除数据
func (m *MovieModel) Delete(id int64) error {
	// 先判断id的基础有效性防止进行不必要的查询
	if id < 1 {
		return ErrRecordNotFound
	}
	stmt := `DELETE FROM movies WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	result, err := m.db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}
	// 获取受影响的行数检查是否删除成功
	rowAffected, err := result.RowsAffected()
	// 如果没有行收到影响则为删除失败
	if rowAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// 根据query url的参数返回指定的信息
func (m *MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	// 使用fmt.Sprintf动态生成查询语句 确保ORDER BY 作用于一个一定存在的key保证输出是有序的
	// psql若没有指定排序输出顺序是随机的
	stmt := fmt.Sprintf(`
		SELECT id,created_at,title,year,runtime,genres,version
		FROM movies
		WHERE (to_tsvector('simple',title) @@ plainto_tsquery('simple',$1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s,id ASC`, filters.sortColumn(), filters.sortDirection())
	// 创建DeadLine
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	// 执行查询请求
	rows, err := m.db.QueryContext(ctx, stmt, title, pq.Array(genres))
	if err != nil {
		return nil, err
	}
	// 读取完毕后回收row的资源
	defer rows.Close()
	// 创建切片用于存储查询到的信息
	movies := []*Movie{}
	// 从rows中提取数据
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}
		// 提取成功将内容加入slice
		movies = append(movies, &movie)
	}
	// 迭代结束检查是否发生错误
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}
