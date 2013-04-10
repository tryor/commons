package dbutil

/*
	condi := new(sqlutil.Condition)
	condi.And(sqlutil.NewConditionItem("c1", "=", 1))
	condi.And(sqlutil.NewConditionItem("c2", "!=", 1.1))
	condi.And(sqlutil.NewConditionItem("c3", ">=", "aaa"))
	condi.And(sqlutil.NewSpecialConditionItem("c4", "between", 2, 4))
	condi.And(sqlutil.NewSpecialConditionItem("c5", "in", 1, 2, 3, 4))

	subcondi := new(sqlutil.Condition)
	subcondi.Or(sqlutil.NewConditionItem("sc1", "=", 1))
	subcondi.Or(sqlutil.NewConditionItem("sc2", "<=", 1.1))
	subcondi.Or(sqlutil.NewConditionItem("sc3", "like", "aaa"))

	condi.And(subcondi)

	fmt.Println("%v", condi.GetStatement())
	fmt.Println("%v", condi.GetValues())

*/

import (
	"fmt"
	"strings"
)

type Condition struct {
	conditions []*term
}

type term struct {
	logicOperator string
	condiItem     *ConditionItem
	subCondi      *Condition
}

//Logical operators, AND
const AND string = "AND"

//Logical operators, OR
const OR string = "OR"

//Logical operators, AND NOT
const AND_NOT string = "AND NOT"

//Logical operators, AND NOT
const OR_NOT string = "OR NOT"

//Build SQL conditions available to all of the logical operation symbols
var LOGIC_OPERATOR_SIGNS []string = []string{
	AND, OR, AND_NOT, OR_NOT}

//Check whether effective logical operators
func isValidLogicOperator(oprtor string) bool {
	for i := 0; i < len(LOGIC_OPERATOR_SIGNS); i++ {
		if LOGIC_OPERATOR_SIGNS[i] == oprtor {
			return true
		}
	}
	return false
}

func (this *Condition) Clear() {
	this.conditions = this.conditions[:0]
}

func (this *Condition) Size() int {
	return len(this.conditions)
}

//Add item condition
//(logicOperator) Logical operators, AND, OR, AND NOT, OR NOT
//(item) Item condition, *ConditionItem or *Condition
func (this *Condition) Add(logicOperator string, item interface{}) *Condition {
	logicOperator = strings.ToUpper(strings.TrimSpace(logicOperator))
	if !isValidLogicOperator(logicOperator) {
		panic(
			"Illegal logical operators, legal is (AND, OR, AND NOT, OR NOT)")
	}
	switch v := item.(type) {
	case *ConditionItem:
		this.conditions = append(this.conditions, &term{logicOperator: logicOperator, condiItem: v})
	case *Condition:
		this.conditions = append(this.conditions, &term{logicOperator: logicOperator, subCondi: v})
	default:
		panic("(item) parameter type error，legal is *ConditionItem or *Condition")
	}
	return this
}

func (this *Condition) And(item interface{}) *Condition {
	return this.Add(AND, item)
}

func (this *Condition) Or(item interface{}) *Condition {
	return this.Add(OR, item)
}

func (this *Condition) AndNot(item interface{}) *Condition {
	return this.Add(AND_NOT, item)
}

func (this *Condition) OrNot(item interface{}) *Condition {
	return this.Add(OR_NOT, item)
}

func (this *Condition) GetStatement() string {
	var wherebuff []interface{}
	var logicOperator string
	for _, trm := range this.conditions {
		logicOperator = trm.logicOperator
		if len(wherebuff) == 0 {
			if logicOperator == AND || logicOperator == OR {
				logicOperator = ""
			} else if logicOperator == AND_NOT || logicOperator == OR_NOT {
				logicOperator = "NOT"
			}
		}

		if trm.condiItem != nil {
			wherebuff = append(wherebuff, logicOperator, " ", trm.condiItem.GetString())
		} else {
			var subCondnStr string = trm.subCondi.GetStatement()
			if len(subCondnStr) > 0 {
				wherebuff = append(wherebuff, logicOperator, " (", subCondnStr, ") ")
			}
		}
	}

	return fmt.Sprint(wherebuff...)
}

func (this *Condition) GetValues() []interface{} {
	var values []interface{}
	for _, trm := range this.conditions {
		if trm.condiItem != nil {
			vals := trm.condiItem.GetValues()
			if vals != nil {
				values = append(values, vals...)
			}
		} else {
			vals := trm.subCondi.GetValues()
			if vals != nil {
				values = append(values, vals...)
			}
		}
	}
	return values
}

func (this *Condition) String() string {
	return fmt.Sprint(this.GetStatement(), this.GetValues())
}
