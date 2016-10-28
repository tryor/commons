/*
Need to set the driver before using，
import (_ "github.com/Go-SQL-Driver/MySQL")
or
import (_ "github.com/ziutek/mymysql/godrv")
and
Driver and DSN
*/
package dbutil

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/trygo/beedb"
)

var (
	Driver, DSN string
)

func OpenDB(param ...string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	if len(param) >= 2 {
		db, err = sql.Open(param[0], param[1])
	} else {
		db, err = sql.Open(Driver, DSN)
	}
	if err != nil {
		//panic(err)
		return nil, err
	}
	return db, nil
}

func CloseDB(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

//In a transaction block of executive function
func Transaction(db *sql.DB, f func(tx *sql.Tx) error) (err error) {

	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return
	}

	//如果f()函数是通过panic抛出错误，那也将此错误使用panic抛出
	defer func() {
		if r := recover(); r != nil {
			err = tx.Rollback()
			if err != nil {
				panic(err)
			}
			panic(r)
		} else if err == nil {
			err = tx.Commit()
		}
	}()

	err = f(tx)
	if err != nil {
		err1 := tx.Rollback()
		if err1 != nil {
			err = err1
		}
	}
	return err
}

//count records
func Count(db *sql.DB, table string, where string, params ...interface{}) (int64, error) {
	if strings.TrimSpace(where) != "" {
		return CountQuery(db, fmt.Sprint("select count(*) from ", table, " where ", where), params...)
	} else {
		return CountQuery(db, "select count(*) from "+table)
	}
}

//count records
func CountQuery(db *sql.DB, sql string, params ...interface{}) (int64, error) {

	if beedb.OnDebug {
		log.Println(sql)
		log.Println(params)
	}

	s, err := db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	defer s.Close()
	res, err := s.Query(params...)
	if err != nil {
		return 0, err
	}
	defer res.Close()

	if res.Next() {
		var countResults []interface{}
		var countResult interface{}
		countResults = append(countResults, &countResult)
		if err := res.Scan(countResults...); err != nil {
			return 0, err
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
		return countVal, nil
	}
	return 0, nil
}

func FindAll(db *sql.DB, rowsSlicePtr interface{}, sql string, params ...interface{}) error {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return errors.New("needs a pointer to a slice")
	}

	sliceElementType := sliceValue.Type().Elem()

	resultsSlice, err := FindMap(db, sql, params...)
	if err != nil {
		return err
	}

	for _, results := range resultsSlice {
		newValue := reflect.New(sliceElementType)
		//err := beedb.ScanMapIntoStruct(newValue.Interface(), results)
		err := ScanMapIntoStruct(newValue.Interface(), results)
		if err != nil {
			return err
		}
		sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
	}
	return nil
}

func FindMap(db *sql.DB, sql string, params ...interface{}) (resultsSlice []map[string][]byte, err error) {

	if beedb.OnDebug {
		log.Println(sql)
		log.Println(params)
	}

	res, err := db.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	fields, err := res.Columns()
	if err != nil {
		return nil, err
	}
	for res.Next() {
		result := make(map[string][]byte)
		var scanResultContainers []interface{}
		for i := 0; i < len(fields); i++ {
			var scanResultContainer interface{}
			scanResultContainers = append(scanResultContainers, &scanResultContainer)
		}
		if err := res.Scan(scanResultContainers...); err != nil {
			return nil, err
		}
		for ii, key := range fields {
			rawValue := reflect.Indirect(reflect.ValueOf(scanResultContainers[ii]))
			//if row is null then ignore
			if rawValue.Interface() == nil {
				continue
			}
			aa := reflect.TypeOf(rawValue.Interface())
			vv := reflect.ValueOf(rawValue.Interface())
			var str string
			switch aa.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				str = strconv.FormatInt(vv.Int(), 10)
				result[key] = []byte(str)
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				str = strconv.FormatUint(vv.Uint(), 10)
				result[key] = []byte(str)
			case reflect.Float32, reflect.Float64:
				str = strconv.FormatFloat(vv.Float(), 'f', -1, 64)
				result[key] = []byte(str)
			case reflect.Slice:
				if aa.Elem().Kind() == reflect.Uint8 {
					result[key] = rawValue.Interface().([]byte)
					break
				}
			case reflect.String:
				str = vv.String()
				result[key] = []byte(str)
			//时间类型
			case reflect.Struct:
				str = rawValue.Interface().(time.Time).Format("2006-01-02 15:04:05.000 -0700")
				result[key] = []byte(str)
			}

		}
		resultsSlice = append(resultsSlice, result)
	}
	return resultsSlice, nil
}

//the structure properties of scanning into the map
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

func titleCasedName(name string) string {
	newstr := make([]rune, 0)
	upNextChar := true

	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			chr -= ('a' - 'A')
		case chr == '_':
			upNextChar = true
			continue
		}

		newstr = append(newstr, chr)
	}

	return string(newstr)
}

func ScanMapIntoStruct(obj interface{}, objMap map[string][]byte) error {
	dataStruct := reflect.Indirect(reflect.ValueOf(obj))
	if dataStruct.Kind() != reflect.Struct {
		return errors.New("expected a pointer to a struct")
	}

	for key, data := range objMap {
		structField := dataStruct.FieldByName(titleCasedName(key))
		if !structField.CanSet() {
			continue
		}

		var v interface{}

		switch structField.Type().Kind() {
		case reflect.Slice:
			v = data
		case reflect.String:
			v = string(data)
		case reflect.Bool:
			v = string(data) == "1"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			x, err := strconv.Atoi(string(data))
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		case reflect.Int64:
			x, err := strconv.ParseInt(string(data), 10, 64)
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(string(data), 64)
			if err != nil {
				return errors.New("arg " + key + " as float64: " + err.Error())
			}
			v = x
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x, err := strconv.ParseUint(string(data), 10, 64)
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		//Now only support Time type
		case reflect.Struct:
			if structField.Type().String() != "time.Time" {
				return errors.New("unsupported struct type in Scan: " + structField.Type().String())
			}

			x, err := time.Parse("2006-01-02 15:04:05", string(data))
			if err != nil {
				x, err = time.Parse("2006-01-02 15:04:05.000 -0700", string(data))

				if err != nil {
					return errors.New("unsupported time format: " + string(data))
				}
			}

			v = x
		default:
			return errors.New("unsupported type in Scan: " + reflect.TypeOf(v).String())
		}

		structField.Set(reflect.ValueOf(v))
	}

	return nil
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
