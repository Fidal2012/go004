package week02

import (
	"database/sql"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

// 个人理解数据访问对象遇到sql.ErrNoRows时应根据业务场景判断是否返回错误
// 1. 类似Select操作遇到ErrNoRows问题只需降级处理，返回 空集，nil 即可，遇到不可处理问题wrap关键字段信息并返回，交由上层处理
// 2. 类似Get/Set操作，在业务中多是需要返回“未找到ID为XXX的用户”，需要wrap关键查询字段并返回

type User struct {
	id   int
	name string
}

func Get(id int) (User, error) {
	err := sql.ErrNoRows
	err = errors.Wrap(err, "Get User where id = "+strconv.Itoa(id))
	return User{}, err
}

func Select(id ...int) ([]User, error) {
	var err, sqlErr error
	result := make([]User, 0, len(id))
	for i := range id {
		// operation
		u := User{id: i}
		if i/2 == 0 {
			sqlErr = sql.ErrNoRows
		} else {
			sqlErr = nil
		}
		time.Sleep(time.Second)
		// operation over, return error

		// handle error
		if sqlErr != nil {
			if sqlErr == sql.ErrNoRows {
				// print log
				continue
			}
			// return first error
			if err == nil {
				err = errors.Wrap(sqlErr, "select User where id = "+strconv.Itoa(i))
			}
		}
		result = append(result, u)
	}
	return result, err
}
