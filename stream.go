package stream_for_go

import (
	"reflect"
	"sort"
	"sync"
)

type Stream interface {
	Filter(predicate Predicate) Stream
	Map(function Function) Stream
	ForEach(consumer Consumer)
	Peek(consumer Consumer) Stream
	Limit(maxSize int) Stream
	Skip(n int) Stream
	Sorted(comparator Comparator) Stream
	Distinct(comparator Comparator) Stream
	AllMatch(predicate Predicate) bool
	AnyMatch(predicate Predicate) bool
	NoneMatch(predicate Predicate) bool
	Count() int
	Reduce(function BiFunction) interface{}
	ToSlice(targetSlice interface{})
}

type TerminalOp interface {
	EvaluateParallel(sourceStage *pipeline)
	EvaluateSequential(sourceStage *pipeline)
}

type Predicate func(v interface{}) bool

type Function func(v interface{}) interface{}

type Consumer func(v interface{})

type Comparator func(i, j interface{}) bool

type BiFunction func(t, u interface{}) interface{}

type sortData struct {
	data       []interface{}
	comparator Comparator
}

func (s *sortData) Len() int {
	return len(s.data)
}
func (s *sortData) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}
func (s *sortData) Less(i, j int) bool {
	return s.comparator(s.data[i], s.data[j])
}

type ForEachOp struct {
}

func (f ForEachOp) EvaluateParallel(sourceStage *pipeline) {
	headStage := sourceStage.nextStage
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(sourceStage.data))
	for _, v := range sourceStage.data {
		data:=v
		go func() {
			defer waitGroup.Done()
			headStage.do(headStage.nextStage, data)
		}()
	}
	waitGroup.Wait()
}

func (f ForEachOp) EvaluateSequential(sourceStage *pipeline) {
	headStage := sourceStage.nextStage
	for _, v := range sourceStage.data {
		headStage.do(headStage.nextStage, v)
		if sourceStage.stop {
			break
		}
	}
}

func Parallel(arr interface{}) *pipeline {
	p := New(arr)
	p.parallel = true
	return p
}

func New(arr interface{}) *pipeline {
	data := make([]interface{}, 0)
	dataValue := reflect.ValueOf(&data).Elem()
	arrValue := reflect.ValueOf(arr)
	if arrValue.Kind() == reflect.Ptr {
		arrValue = arrValue.Elem()
	}
	if arrValue.Kind() == reflect.Slice || arrValue.Kind() == reflect.Array {
		for i := 0; i < arrValue.Len(); i++ {
			dataValue.Set(reflect.Append(dataValue, arrValue.Index(i)))
		}
	} else {
		panic("the type of arr parameter must be Array or Slice")
	}
	p := &pipeline{data: data}
	p.sourceStage = p
	return p
}

type pipeline struct {
	lock           sync.Mutex
	data, tmpData  []interface{}
	sourceStage    *pipeline
	previousStage  *pipeline
	nextStage      *pipeline
	parallel, stop bool
	do             func(nextStage *pipeline, v interface{})
}

func (p *pipeline) ToSlice(targetSlice interface{}) {
	targetValue := reflect.ValueOf(targetSlice)
	if targetValue.Kind() != reflect.Ptr {
		panic("target slice must be a pointer")
	}
	sliceValue := reflect.Indirect(targetValue)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(v)))
		},
	}
	t.evaluate(&ForEachOp{})
}

func (p *pipeline) Reduce(function BiFunction) interface{} {
	p.nilCheck(function)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			if p.tmpData == nil {
				p.tmpData = make([]interface{}, len(p.sourceStage.data))
			}
			if p.tmpData[0] == nil {
				p.tmpData[0] = v
			} else {
				res := function(p.tmpData[0], v)
				p.tmpData[0] = res
			}
		},
	}
	t.evaluate(&ForEachOp{})
	return p.tmpData[0]
}

func (p *pipeline) Count() int {
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	return len(p.tmpData)
}

func (p *pipeline) NoneMatch(predicate Predicate) bool {
	return !p.AnyMatch(predicate)
}

func (p *pipeline) AnyMatch(predicate Predicate) bool {
	return p.matchOps(predicate, true)
}

func (p *pipeline) AllMatch(predicate Predicate) bool {
	return p.matchOps(predicate, false)
}

func (p *pipeline) matchOps(predicate Predicate, flag bool) bool {
	p.nilCheck(predicate)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			match := predicate(v)
			if !flag {
				match = !match
			}
			if match {
				p.sourceStage.stop = true
			}
		},
	}
	t.evaluate(&ForEachOp{})
	if !flag {
		return !p.sourceStage.stop
	}
	return p.sourceStage.stop
}

func (p *pipeline) Distinct(comparator Comparator) Stream {
	p.nilCheck(comparator)
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	for i := range p.tmpData {
		flag := true
		for j := range t.data {
			if comparator(p.tmpData[i], t.data[j]) {
				flag = false
				break
			}
		}
		if flag {
			t.data = append(t.data, p.tmpData[i])
		}
	}
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Sorted(comparator Comparator) Stream {
	p.nilCheck(comparator)
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	s := &sortData{data: p.tmpData, comparator: comparator}
	sort.Sort(s)
	t.data = p.tmpData
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Skip(n int) Stream {
	if n < 0 {
		n = 0
	}
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	dataLen := len(p.tmpData)
	if dataLen < n {
		n = dataLen
	}
	t.data = p.tmpData[n:]
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Limit(maxSize int) Stream {
	if maxSize < 0 {
		maxSize = 0
	}
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	dataLen := len(p.tmpData)
	if dataLen < maxSize {
		maxSize = dataLen
	}
	t.data = p.tmpData[:maxSize]
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Peek(consumer Consumer) Stream {
	p.nilCheck(consumer)
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			consumer(v)
			nextStage.do(nextStage.nextStage, v)
		},
	}
}

func (p *pipeline) Filter(predicate Predicate) Stream {
	p.nilCheck(predicate)
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if predicate(v) {
				nextStage.do(nextStage.nextStage, v)
			}
		},
	}
}

func (p *pipeline) ForEach(consumer Consumer) {
	p.nilCheck(consumer)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			consumer(v)
		},
	}
	t.evaluate(&ForEachOp{})
}

func (p *pipeline) Map(function Function) Stream {
	p.nilCheck(function)
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			nextStage.do(nextStage.nextStage, function(v))
		},
	}
}

func (p *pipeline) evaluate(op TerminalOp) {
	p.nilCheck(op)
	for headStage := p; headStage != nil && headStage.previousStage != nil; headStage = headStage.previousStage {
		headStage.previousStage.nextStage = headStage
	}

	if p.sourceStage.parallel {
		op.EvaluateParallel(p.sourceStage)
	} else {
		op.EvaluateSequential(p.sourceStage)
	}
}

func (p *pipeline) nilCheck(v interface{}) {
	if v == nil {
		panic("nil forbidden")
	}
}

func (p *pipeline) statefulStage() *pipeline {
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			p.addTmpData(v)
		},
	}
}

func (p *pipeline) addTmpData(v interface{}) {
	if p.parallel {
		p.lock.Lock()
		defer p.lock.Unlock()
	}
	p.tmpData = append(p.tmpData, v)
}
