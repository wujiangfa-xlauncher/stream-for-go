package stream_for_go

import (
	"fmt"
	"math/rand"
	"testing"
)

type student struct {
	id     int
	name   string
	age    int
	scores []int
}

func (s *student) String() string {
	return fmt.Sprintf("{id:%d, name:%s, age:%d,scores:%v}", s.id, s.name, s.age, s.scores)
}

func createStudents() []student {
	names := []string{"Tom", "Kate", "Lucy", "Jim", "Jack", "King", "Lee", "Mask"}
	students := make([]student, 10)
	rnd := func(start, end int) int { return rand.Intn(end-start) + start }
	for i := 0; i < 10; i++ {
		students[i] = student{
			id:     i + 1,
			name:   names[rand.Intn(len(names))],
			age:    rnd(15, 26),
			scores: []int{rnd(60, 100), rnd(60, 100), rnd(60, 100)},
		}
	}
	return students
}

func Test1(t *testing.T) {

	students := createStudents()
	New(students).Peek(func(v interface{}) {
		s := v.(student)
		fmt.Println(s.String())
	}).Filter(func(v interface{}) bool {
		s := v.(student)
		return len(s.name) > 3
	}).Peek(func(v interface{}) {
		s := v.(student)
		fmt.Println(s.String())
	}).Map(func(v interface{}) interface{} {
		s := v.(student)
		return s.name
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func Test2(t *testing.T) {
	students := createStudents()
	New(students).Sorted(func(i, j interface{}) bool {
		return i.(student).age < j.(student).age
	}).Skip(1).Distinct(func(i, j interface{}) bool {
		return i.(student).age == j.(student).age
	}).Map(func(v interface{}) interface{} {
		return v.(student).name
	}).Distinct(func(i, j interface{}) bool {
		return i == j
	}).Filter(func(v interface{}) bool {
		return len(v.(string)) > 3
	}).Map(func(v interface{}) interface{} {
		return v.(string) + "01"
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func Test3(t *testing.T) {
	students := createStudents()
	allMatch := New(students).Peek(func(v interface{}) {
		fmt.Println(v.(student).name)
	}).Map(func(v interface{}) interface{} {
		return v.(student).name
	}).NoneMatch(func(v interface{}) bool {
		return len(v.(string)) >= 5
	})
	fmt.Println(allMatch)
}

func Test4(t *testing.T) {
	students := createStudents()
	count := New(students).Map(func(v interface{}) interface{} {
		return v.(student).age
	}).Sorted(func(i, j interface{}) bool {
		return i.(int) > j.(int)
	}).Distinct(func(i, j interface{}) bool {
		return i == j
	}).Filter(func(v interface{}) bool {
		return v.(int) > 0
	}).Peek(func(v interface{}) {
		fmt.Println(v)
	}).Count()
	fmt.Println(count)
}

func Test5(t *testing.T) {
	students := createStudents()
	reduce := New(students).Peek(func(v interface{}) {
		fmt.Println(v)
	}).Map(func(v interface{}) interface{} {
		return v.(student).name
	}).Reduce(func(t, u interface{}) interface{} {
		return t.(string) + "," + u.(string)
	})
	fmt.Println(reduce)
}

func Test6(t *testing.T) {
	students := createStudents()
	var data []student
	New(students).Filter(func(v interface{}) bool {
		return v.(student).age > 20
	}).ToSlice(&data)
	fmt.Println(data)
}

func Test7(t *testing.T) {
	students := createStudents()
	reduce := Parallel(students).Peek(func(v interface{}) {
		fmt.Println(v)
	}).Map(func(v interface{}) interface{} {
		return v.(student).age
	}).Reduce(func(t, u interface{}) interface{} {
		return t.(int) + u.(int)
	})
	fmt.Println(reduce)
}

func Test8(t *testing.T) {
	students := createStudents()
	match := Parallel(students).Peek(func(v interface{}) {
		s := v.(student)
		fmt.Println(s.String())
	}).NoneMatch(func(v interface{}) bool {
		return v.(student).age > 30
	})
	fmt.Println(match)
}

func Test9(t *testing.T) {
	students := createStudents()
	count := Parallel(students).Map(func(v interface{}) interface{} {
		return v.(student).age
	}).Sorted(func(i, j interface{}) bool {
		return i.(int) > j.(int)
	}).Distinct(func(i, j interface{}) bool {
		return i == j
	}).Filter(func(v interface{}) bool {
		return v.(int) > 0
	}).Peek(func(v interface{}) {
		fmt.Println(v)
	}).Count()
	fmt.Println(count)
}
