package stream

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

func TestForEach(t *testing.T) {
	students := createStudents()
	New(students).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestFilter(t *testing.T) {
	students := createStudents()
	New(students).Filter(func(v interface{}) bool {
		return v.(student).age>20
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestMap(t *testing.T) {
	students := createStudents()
	New(students).Map(func(v interface{}) interface{} {
		return v.(student).name
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})

	fmt.Println("--------------")

	var names = [4]string{"zhangsan","lisi","wangwu","zhaoliu"}
	New(names).Map(func(v interface{}) interface{} {
		s := v.(string)
		return student{
			id: len(s),
			name:   s,
			age:    len(s)*4,
		}
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestPeek(t *testing.T) {
	students := createStudents()
	New(students).Peek(func(v interface{}) {
		fmt.Println(v.(student).scores)
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestStateless(t *testing.T) {
	students := createStudents()
	New(students).Peek(func(v interface{}) {
		fmt.Println(v)
	}).Filter(func(v interface{}) bool {
		return v.(student).age>20
	}).Map(func(v interface{}) interface{} {
		return v.(student).name
	}).Filter(func(v interface{}) bool {
		return len(v.(string))>3
	}).ForEach(func(v interface{}) {
		fmt.Println("Res:"+v.(string))
	})
}

func TestSkip(t *testing.T) {
	students := createStudents()
	New(students).Skip(5).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestLimit(t *testing.T) {
	students := createStudents()
	New(students).Limit(5).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestDistinct(t *testing.T) {
	students := createStudents()
	New(students).Distinct(func(i, j interface{}) bool {
		return i.(student).name == j.(student).name
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestSorted(t *testing.T) {
	students := createStudents()
	New(students).Sorted(func(i, j interface{}) bool {
		return i.(student).age < j.(student).age
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestStateful(t *testing.T) {
	students := createStudents()
	New(students).Limit(7).Distinct(func(i, j interface{}) bool {
		return i.(student).name == j.(student).name
	}).Sorted(func(i, j interface{}) bool {
		return i.(student).age < j.(student).age
	}).ForEach(func(v interface{}) {
		fmt.Println(v)
	})
}

func TestAllMatch(t *testing.T) {
	students := createStudents()
	allMatch := New(students).Peek(func(v interface{}) {
		fmt.Println(v)
	}).AllMatch(func(v interface{}) bool {
		return v.(student).age > 15
	})
	fmt.Println(allMatch)
}

func TestAnyMatch(t *testing.T) {
	students := createStudents()
	anyMatch := New(students).Peek(func(v interface{}) {
		fmt.Println(v)
	}).AnyMatch(func(v interface{}) bool {
		return v.(student).age > 20
	})
	fmt.Println(anyMatch)
}

func TestNoneMatch(t *testing.T) {
	students := createStudents()
	noneMatch := New(students).Peek(func(v interface{}) {
		fmt.Println(v)
	}).NoneMatch(func(v interface{}) bool {
		return v.(student).age > 20
	})
	fmt.Println(noneMatch)
}

func TestCount(t *testing.T) {
	students := createStudents()
	count := New(students).Count()
	fmt.Println(count)
	filterCount := New(students).Filter(func(v interface{}) bool {
		return v.(student).age > 20
	}).Count()
	fmt.Println(filterCount)
}

func TestReduce(t *testing.T) {
	students := createStudents()
	name := New(students).Map(func(v interface{}) interface{} {
		return v.(student).name
	}).Reduce(func(t, u interface{}) interface{} {
		return t.(string) + "," + u.(string)
	})
	fmt.Println(name)

	age := New(students).Map(func(v interface{}) interface{} {
		return v.(student).age
	}).Reduce(func(t, u interface{}) interface{} {
		return t.(int) + u.(int)
	})
	fmt.Println(age)
}

func TestToSlice(t *testing.T) {
	students := createStudents()
	var ageArray []int
	New(students).Map(func(v interface{}) interface{} {
		return v.(student).age
	}).ToSlice(&ageArray)
	fmt.Println(ageArray)

	var nameArray []string
	New(students).Map(func(v interface{}) interface{} {
		return v.(student).name
	}).ToSlice(&nameArray)
	fmt.Println(nameArray)

	var studentArray []student
	New(students).Filter(func(v interface{}) bool {
		return len(v.(student).name) > 3
	}).ToSlice(&studentArray)
	fmt.Println(studentArray)
}

func TestStream(t *testing.T) {
	students := createStudents()
	count := New(students).Map(func(v interface{}) interface{} {
		return v.(student).age
	}).Sorted(func(i, j interface{}) bool {
		return i.(int) > j.(int)
	}).Distinct(func(i, j interface{}) bool {
		return i == j
	}).Filter(func(v interface{}) bool {
		return v.(int) > 16
	}).Peek(func(v interface{}) {
		fmt.Println(v)
	}).Count()
	fmt.Println(count)
}

func TestParallel(t *testing.T) {
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
