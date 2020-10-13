# stream-for-go
Implementation of Java stream API by go language

----------
**Usage**
```
import (
	"fmt"
	"github.com/wujiangfa-xlauncher/stream-for-go"
)

func main() {
	var ints = []int{1, 3, 4, 5, -2, 1, 2, 3, 5, 7, 10, -6}
	reduce := stream.New(ints).Filter(func(v interface{}) bool {
		return v.(int) > 0
	}).Distinct(func(i, j interface{}) bool {
		return i == j
	}).Peek(func(v interface{}) {
		fmt.Println(v)
	}).Reduce(func(t, u interface{}) interface{} {
		return t.(int) + u.(int)
	})
	fmt.Println(reduce)
}
```
----------

**Demo Preparation**

    type student struct {
    	id int
    	name   string
    	ageint
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
    			id: i + 1,
    			name:   names[rand.Intn(len(names))],
    			age:rnd(15, 26),
    			scores: []int{rnd(60, 100), rnd(60, 100), rnd(60, 100)},
    		}
    	}
    	return students
    }
    
----------
### Stateless api
----------
Filter:
```
students := createStudents()
New(students).Filter(func(v interface{}) bool {
	return v.(student).age>20
}).ForEach(func(v interface{}) {
	fmt.Println(v)
})
```
Output:
```
{2 Lee 22 [80 76 80]}
{4 Lucy 22 [65 97 86]}
{7 King 22 [87 91 89]}
{9 King 21 [94 63 93]}
```
Map:
```
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
```
Output:
```
Kate
Lee
Lee
Lucy
Mask
Jim
King
Jack
King
Jim
--------------
{8 zhangsan 32 []}
{4 lisi 16 []}
{6 wangwu 24 []}
{7 zhaoliu 28 []}
```
Peek:
```
students := createStudents()
New(students).Peek(func(v interface{}) {
	fmt.Println(v.(student).scores)
}).ForEach(func(v interface{}) {
	fmt.Println(v)
})
```
Output:
```
[67 79 61]
{1 Kate 16 [67 79 61]}
[80 76 80]
{2 Lee 22 [80 76 80]}
[62 69 68]
{3 Lee 15 [62 69 68]}
[65 97 86]
{4 Lucy 22 [65 97 86]}
[68 78 67]
{5 Mask 15 [68 78 67]}
[68 90 75]
{6 Jim 20 [68 90 75]}
[87 91 89]
{7 King 22 [87 91 89]}
[91 65 86]
{8 Jack 16 [91 65 86]}
[94 63 93]
{9 King 21 [94 63 93]}
[64 99 93]
{10 Jim 20 [64 99 93]}
```
Demo: find students which age > 20 and len(name) >3 
```
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
```
Output:
```
{1 Kate 16 [67 79 61]}
{2 Lee 22 [80 76 80]}
{3 Lee 15 [62 69 68]}
{4 Lucy 22 [65 97 86]}
Res:Lucy
{5 Mask 15 [68 78 67]}
{6 Jim 20 [68 90 75]}
{7 King 22 [87 91 89]}
Res:King
{8 Jack 16 [91 65 86]}
{9 King 21 [94 63 93]}
Res:King
{10 Jim 20 [64 99 93]}
```

----------
### Stateful api
----------
Skip:
```
students := createStudents()
New(students).Skip(5).ForEach(func(v interface{}) {
	fmt.Println(v)
})
```
Output:
```
{6 Jim 20 [68 90 75]}
{7 King 22 [87 91 89]}
{8 Jack 16 [91 65 86]}
{9 King 21 [94 63 93]}
{10 Jim 20 [64 99 93]}
```
Limit:
```
students := createStudents()
New(students).Limit(5).ForEach(func(v interface{}) {
	fmt.Println(v)
})
```
Output:
```
{1 Kate 16 [67 79 61]}
{2 Lee 22 [80 76 80]}
{3 Lee 15 [62 69 68]}
{4 Lucy 22 [65 97 86]}
{5 Mask 15 [68 78 67]}
```
Distinct:
```
students := createStudents()
New(students).Distinct(func(i, j interface{}) bool {
	return i.(student).name == j.(student).name
}).ForEach(func(v interface{}) {
	fmt.Println(v)
})
```
Output:
```
{1 Kate 16 [67 79 61]}
{2 Lee 22 [80 76 80]}
{4 Lucy 22 [65 97 86]}
{5 Mask 15 [68 78 67]}
{6 Jim 20 [68 90 75]}
{7 King 22 [87 91 89]}
{8 Jack 16 [91 65 86]}
```
Sorted:
```
students := createStudents()
New(students).Sorted(func(i, j interface{}) bool {
	return i.(student).age < j.(student).age
}).ForEach(func(v interface{}) {
	fmt.Println(v)
})
```
Output:
```
{3 Lee 15 [62 69 68]}
{5 Mask 15 [68 78 67]}
{1 Kate 16 [67 79 61]}
{8 Jack 16 [91 65 86]}
{10 Jim 20 [64 99 93]}
{6 Jim 20 [68 90 75]}
{9 King 21 [94 63 93]}
{7 King 22 [87 91 89]}
{2 Lee 22 [80 76 80]}
{4 Lucy 22 [65 97 86]}
```
Demo: distinct and sorted students
```
students := createStudents()
New(students).Limit(7).Distinct(func(i, j interface{}) bool {
    return i.(student).name == j.(student).name
}).Sorted(func(i, j interface{}) bool {
    return i.(student).age < j.(student).age
}).ForEach(func(v interface{}) {
    fmt.Println(v)
})
```
Output:
```
{5 Mask 15 [68 78 67]}
{1 Kate 16 [67 79 61]}
{6 Jim 20 [68 90 75]}
{2 Lee 22 [80 76 80]}
{4 Lucy 22 [65 97 86]}
{7 King 22 [87 91 89]}
```

----------
### Terminal api
----------
ForEach:
```
students := createStudents()
New(students).ForEach(func(v interface{}) {
	fmt.Println(v)
})
```
Output:
```
{1 Kate 16 [67 79 61]}
{2 Lee 22 [80 76 80]}
{3 Lee 15 [62 69 68]}
{4 Lucy 22 [65 97 86]}
{5 Mask 15 [68 78 67]}
{6 Jim 20 [68 90 75]}
{7 King 22 [87 91 89]}
{8 Jack 16 [91 65 86]}
{9 King 21 [94 63 93]}
{10 Jim 20 [64 99 93]}
```
AllMatch:
```
students := createStudents()
allMatch := New(students).Peek(func(v interface{}) {
    fmt.Println(v)
}).AllMatch(func(v interface{}) bool {
    return v.(student).age > 15
})
fmt.Println(allMatch)
```
Output:
```
{1 Kate 16 [67 79 61]}
{2 Lee 22 [80 76 80]}
{3 Lee 15 [62 69 68]}
false
```
AnyMatch:
```
students := createStudents()
anyMatch := New(students).Peek(func(v interface{}) {
    fmt.Println(v)
}).AnyMatch(func(v interface{}) bool {
    return v.(student).age > 20
})
fmt.Println(anyMatch)
```
Output:
```
{1 Kate 16 [67 79 61]}
{2 Lee 22 [80 76 80]}
true
```
NoneMatch:
```
students := createStudents()
noneMatch := New(students).Peek(func(v interface{}) {
    fmt.Println(v)
}).NoneMatch(func(v interface{}) bool {
    return v.(student).age > 20
})
fmt.Println(noneMatch)
```
Output:
```
{1 Kate 16 [67 79 61]}
{2 Lee 22 [80 76 80]}
false
```
Count:
```
students := createStudents()
count := New(students).Count()
fmt.Println(count)
filterCount := New(students).Filter(func(v interface{}) bool {
    return v.(student).age > 20
}).Count()
fmt.Println(filterCount)
```
Output:
```
10
4
```
Reduce:
```
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
```
Output:
```
Kate,Lee,Lee,Lucy,Mask,Jim,King,Jack,King,Jim
189
```
ToSlice:
```
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
```
Output:
```
[16 22 15 22 15 20 22 16 21 20]
[Kate Lee Lee Lucy Mask Jim King Jack King Jim]
[{1 Kate 16 [67 79 61]} {4 Lucy 22 [65 97 86]} {5 Mask 15 [68 78 67]} {7 King 22 [87 91 89]} {8 Jack 16 [91 65 86]} {9 King 21 [94 63 93]}]
```
MaxMin:
```
students := createStudents()
max := New(students).MaxMin(func(i, j interface{}) bool {
    return i.(student).age > j.(student).age
})
fmt.Println(max)

min := Parallel(students).MaxMin(func(i, j interface{}) bool {
    return i.(student).age < j.(student).age
})
fmt.Println(min)

var ints = [10]int{1,3,7,2,6,5,0,-1,-6,-9}
max = New(ints).Peek(func(v interface{}) {
    fmt.Println(v)
}).MaxMin(func(i, j interface{}) bool {
    return i.(int) > j.(int)
})
fmt.Print("max :")
fmt.Println(max)

min = Parallel(ints).Peek(func(v interface{}) {
    fmt.Println(v)
}).MaxMin(func(i, j interface{}) bool {
    return i.(int) < j.(int)
})
fmt.Print("min :")
fmt.Println(min)
```
Output:
```
{7 King 22 [87 91 89]}
{5 Mask 15 [68 78 67]}
1
3
7
2
6
5
0
-1
-6
-9
max :7
-9
5
0
-1
-6
3
7
2
1
6
min :-9
```

----------
### Demo 
----------
Stream demo:
```
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
```
Output:
```
22
21
20
3
```
Parallel demo:
```
students := createStudents()
reduce := Parallel(students).Peek(func(v interface{}) {
    fmt.Println(v)
}).Map(func(v interface{}) interface{} {
    return v.(student).age
}).Reduce(func(t, u interface{}) interface{} {
    return t.(int) + u.(int)
})
fmt.Println(reduce)
```
Output:
```
{10 Jim 20 [64 99 93]}
{4 Lucy 22 [65 97 86]}
{3 Lee 15 [62 69 68]}
{7 King 22 [87 91 89]}
{6 Jim 20 [68 90 75]}
{1 Kate 16 [67 79 61]}
{8 Jack 16 [91 65 86]}
{9 King 21 [94 63 93]}
{2 Lee 22 [80 76 80]}
{5 Mask 15 [68 78 67]}
189
```