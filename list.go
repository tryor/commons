package util

type List struct {
	els []interface{}
}

func NewList(size int, els ...interface{}) *List {
	l := &List{}
	l.els = make([]interface{}, 0, size)
	l.Adds(els...)
	return l
}

func (this *List) Size() int {
	return len(this.els)
}

func (this *List) Adds(els ...interface{}) {
	if len(els) > 0 {
		this.els = append(this.els, els...)
	}
}

func (this *List) Add(el interface{}) {
	this.els = append(this.els, el)
}

func (this *List) Insert(index int, el interface{}) {
	//	rear := this.els[index:]
	//	ss := append(this.els[0:index], el)
	//	this.els = append(ss, rear...)
	this.els = SliceInsert(this.els, index, el)
}

func (this *List) Remove(index int) {
	//this.els = append(this.els[:index], this.els[index+1:]...)
	this.els = SliceRemove(this.els, index, index+1)
}

func (this *List) Gets() []interface{} {
	return this.els
}

func (this *List) Get(index int) interface{} {
	return this.els[index]
}

//如果没找到，返回-1
func (this *List) GetIndex(el interface{}) int {
	for idx, e := range this.els {
		if e == el {
			return idx
		}
	}
	return -1
}

//func InsertOne(slice []interface{}, insertion interface{}, index int) []interface{} {
//	result := make([]interface{}, len(slice)+1)
//	at := copy(result, slice[:index])
//	at += copy(result[at:], insertion)
//	copy(result[at:], slice[index:])
//	return result
//}

func SliceInsert(slice []interface{}, index int, insertion ...interface{}) []interface{} {
	result := make([]interface{}, len(slice)+len(insertion))
	at := copy(result, slice[:index])
	at += copy(result[at:], insertion)
	copy(result[at:], slice[index:])
	return result
}

func SliceRemove(slice []interface{}, start, end int) []interface{} {
	result := make([]interface{}, len(slice)-(end-start))
	at := copy(result, slice[:start])
	copy(result[at:], slice[end:])
	return result
}
