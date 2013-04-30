package dbutil

import (
	"database/sql"
	"errors"
	"fmt"
	//_ "github.com/Go-SQL-Driver/MySQL"
	//_ "github.com/ziutek/mymysql/godrv"
	"reflect"
)

var (
	Driver, DSN string
)

func OpenDB(param ...string) *sql.DB {
	var db *sql.DB
	var err error
	if len(param) >= 2 {
		db, err = sql.Open(param[0], param[1])
	} else {
		db, err = sql.Open(Driver, DSN)
	}
	if err != nil {
		panic(err)
	}
	return db
}

func CloseDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

func Transaction(db *sql.DB, f func()) {

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	defer func() {
		if r := recover(); r != nil {
			err := tx.Rollback()
			if err != nil {
				panic(err)
			}
			panic(r)
		} else {
			err = tx.Commit()
			if err != nil {
				panic(err)
			}
		}
	}()
	f()
}

func Count(db *sql.DB, table string, where string, params ...interface{}) int64 {
	return CountQuery(db, fmt.Sprint("select count(*) from ", table, " where ", where), params...)
}

func CountQuery(db *sql.DB, sql string, params ...interface{}) int64 {
	s, err := db.Prepare(sql)
	if err != nil {
		panic(err)
	}
	defer s.Close()
	res, err := s.Query(params...)
	if err != nil {
		panic(err)
	}
	defer res.Close()

	if res.Next() {
		var countResults []interface{}
		var countResult interface{}
		countResults = append(countResults, &countResult)
		if err := res.Scan(countResults...); err != nil {
			panic(err)
		}
		rawValue := reflect.Indirect(reflect.ValueOf(countResult))
		aa := reflect.TypeOf(rawValue.Interface())
		vv := reflect.ValueOf(rawValue.Interface())
		var countVal int64 = 0
		switch aa.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			countVal = vv.Int()
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			countVal = int64(vv.Uint())
		}
		return countVal
	}
	return 0
}

func ScanStructIntoMap(obj interface{}) (map[string]interface{}, error) {
	dataStruct := reflect.Indirect(reflect.ValueOf(obj))
	if dataStruct.Kind() != reflect.Struct {
		return nil, errors.New("expected a pointer to a struct")
	}

	dataStructType := dataStruct.Type()

	mapped := make(map[string]interface{})

	for i := 0; i < dataStructType.NumField(); i++ {
		field := dataStructType.Field(i)
		fieldName := field.Name

		mapKey := snakeCasedName(fieldName)
		value := dataStruct.FieldByName(fieldName).Interface()

		mapped[mapKey] = value
	}

	return mapped, nil
}

func snakeCasedName(name string) string {
	newstr := make([]rune, 0)
	firstTime := true

	for _, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if firstTime == true {
				firstTime = false
			} else {
				newstr = append(newstr, '_')
			}
			chr -= ('A' - 'a')
		}
		newstr = append(newstr, chr)
	}

	return string(newstr)
}

func IsNoRecord(err error) bool {
	return err != nil && "No record found" == err.Error()
}
