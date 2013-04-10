package dbutil

import (
	"fmt"
	"strings"
)

type ConditionItem struct {
	name     string
	operator string
	values   []interface{}
	isquote  bool // default value is false
	sql      *SQL
}

//Build SQL conditions used when computing symbols: "=", ">=", "<=", ">", "<", "<>", "!=", "LIKE", "NOT LIKE" ,"IS"
var OPERATOR_SIGNS_SINGLE = []string{
	"=", ">=", "<=", ">", "<", "<>", "!=", "LIKE", "NOT LIKE", "IS"}

//Build SQL conditions used when computing symbols: "Between","NOT Between", "IN", "NOT IN" 
var OPERATOR_SIGNS_MORE = []string{
	"BETWEEN", "NOT BETWEEN", "IN", "NOT IN"}

//Build SQL conditions used when computing symbols: "Between","NOT Between"
var OPERATOR_SIGNS_BETWEEN = []string{
	"BETWEEN", "NOT BETWEEN"}

//Build SQL conditions used when computing symbols: "IN", "NOT IN"
var OPERATOR_SIGNS_IN = []string{
	"IN", "NOT IN"}

/**
 * 构造方法, 创建条件单项
 * 
 * @param name 列名称
 * @param operator 条件运算符,合法的运算符有: "=", ">=", "<=", ">", "<", "<>", "!=", "LIKE", "NOT LIKE","IS"
 * @param value 条件值,可以是串类型,或其它能正确转换为数据库数据类型的类型
 * @param isquote 是否是直接引用条件值, 如果是，将会把整个条件项构建为SQL串，而不会将值用？号替换
 */
func NewConditionItem(name string, operator string, value interface{}, isquote ...bool) *ConditionItem {
	operator = strings.ToUpper(strings.TrimSpace(operator))
	if !isValidOperator(OPERATOR_SIGNS_SINGLE, operator) {
		panic("Illegal conditional operator, the correct is (=, >=, <=, >, <, <>, !=, LIKE, NOT LIKE, IS)")
	}
	item := new(ConditionItem)
	item.name = name
	item.operator = operator
	item.values = []interface{}{value}
	if operator == "IS" {
		item.isquote = true
		v := strings.ToUpper(strings.TrimSpace(value.(string)))
		if !(v == "NULL" || v == "NOT NULL") {
			panic("If the operator IS, condition value can only be: NULL or NOT NULL")
		}
	} else {
		if len(isquote) > 0 {
			item.isquote = isquote[0]
		} else {
			item.isquote = false
		}
	}
	return item
}

/**
 * 构造方法, 创建条件单项
 * 
 * @param name 列名称
 * @param operator 条件运算符,合法的运算符有: "=", ">=", "<=", ">", "<", "<>", "!=", "LIKE", "NOT LIKE"
 * @param sql 条件值, 此条件值本身就是一个SQL语句,这样可以在条件中构建子查询
 */
func NewSQLConditionItem(name string, operator string, sql *SQL) *ConditionItem {
	operator = strings.ToUpper(strings.TrimSpace(operator))
	if !isValidOperator(OPERATOR_SIGNS_SINGLE, operator) && !isValidOperator(OPERATOR_SIGNS_IN, operator) {
		panic(
			"Illegal conditional operator, the correct is (=, >=, <=, >, <, <>, !=, LIKE, NOT LIKE, IN, NOT IN)")
	}
	item := new(ConditionItem)
	item.name = name
	item.operator = operator
	item.sql = sql
	return item
}

/**
 * 构造方法, 创建条件单项
 * 
 * @param name 列名称
 * @param operator 条件运算符,合法的运算符有: Between,NOT Between, IN, NOT IN
 * @param values 条件值数组,数组值可以是串类型,或其它能正确转换为数据库数据类型的JAVA类类型,
 *               注意:数据组的长度与条件运算符operator有一定关系,如果运算符是Between,
 *               values数组的值必须是2个,如果运算符是IN,values数组的值至少有1个
 */
func NewSpecialConditionItem(name string, operator string, values ...interface{}) *ConditionItem {
	if len(values) == 0 {
		panic("Conditional value cannot be empty")
	}
	operator = strings.ToUpper(strings.TrimSpace(operator))
	if !isValidOperator(OPERATOR_SIGNS_MORE, operator) {
		panic("Illegal conditional operator, the correct is (Between,NOT Between, IN, NOT IN)")
	}

	if isValidOperator(OPERATOR_SIGNS_BETWEEN, operator) {
		if len(values) != 2 {
			panic("Condition number is not correct，Between, Not Between operation conditions of value number must be of 2")
		}
	}
	item := new(ConditionItem)
	item.name = name
	item.operator = operator
	item.values = values
	return item
}

func (this *ConditionItem) GetValues() []interface{} {
	if this.isquote {
		return nil
	}
	if this.sql != nil {
		this.values = this.sql.Values
	}
	return this.values
}

func (this *ConditionItem) GetString() string {
	if this.sql != nil {
		return fmt.Sprint(this.name, " ", this.operator, " (", this.sql.String, ") ")
	}
	if isValidOperator(OPERATOR_SIGNS_SINGLE, this.operator) {
		if this.isquote {
			return fmt.Sprint(this.name, " ", this.operator, " ", this.values[0], " ")
		} else {
			return fmt.Sprint(this.name, " ", this.operator, " ? ")
		}
	} else {
		if isValidOperator(OPERATOR_SIGNS_BETWEEN, this.operator) {
			return fmt.Sprint(this.name, " ", this.operator, " ? AND ? ")
		} else {
			var valstrs []interface{}
			valstrs = append(valstrs, this.name, " ", this.operator, " (")
			for i := 0; i < len(this.values); i++ {
				if i == 0 {
					valstrs = append(valstrs, "? ")
				} else {
					valstrs = append(valstrs, ",? ")
				}
			}
			valstrs = append(valstrs, ") ")
			return fmt.Sprint(valstrs...)
		}
	}
	return ""
}

/**
 * 检索是否有效的条件运算符
 * 
 * @param oprtor 条件运算符号
 * 
 * @return true,是, false,否
 */
func isValidOperator(definedSigns []string, oprtor string) bool {
	for _, s := range definedSigns {
		if s == oprtor {
			return true
		}
	}
	return false
}
