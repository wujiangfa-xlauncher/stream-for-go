package stream

import (
	"reflect"
	"sort"
	"sync"
)

type Stream interface {
	Filter(predicate Predicate) Stream
	Map(function Function) Stream
	FlatMap(function Function) Stream
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
	MaxMin(comparator Comparator) interface{}
	FindFirst(predicate Predicate) interface{}
	Group(function Function) map[interface{}][]interface{}
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
		data := v
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

func Parallel(arr interface{}) Stream {
	return stream(arr, true)
}

func New(arr interface{}) Stream {
	return stream(arr, false)
}

func stream(arr interface{}, parallel bool) Stream {
	nilCheck(arr)
	data := make([]interface{}, 0)
	dataValue := reflect.ValueOf(&data).Elem()
	arrValue := reflect.ValueOf(arr)
	kindCheck(arrValue)
	for i := 0; i < arrValue.Len(); i++ {
		dataValue.Set(reflect.Append(dataValue, arrValue.Index(i)))
	}
	p := &pipeline{data: data, parallel: parallel}
	p.sourceStage = p
	return p
}

var _ Stream = &pipeline{}

type pipeline struct {
	lock                    sync.Mutex
	data, tmpData           []interface{}
	previousStage           *pipeline
	sourceStage             *pipeline
	nextStage               *pipeline
	parallel, entered, stop bool
	do                      func(nextStage *pipeline, v interface{})
}

func (p *pipeline) Group(function Function) map[interface{}][]interface{} {
	nilCheck(function)
	res := make(map[interface{}][]interface{})
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			out := function(v)
			if out != nil {
				if value, ok := res[out]; ok {
					value = append(value, v)
					res[out] = value
				} else {
					res[out] = []interface{}{v}
				}
			}
		},
	}
	t.evaluate(&ForEachOp{})
	return res
}

func (p *pipeline) FlatMap(function Function) Stream {
	nilCheck(function)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			out := function(v)
			if out != nil {
				data := make([]interface{}, 0)
				dataValue := reflect.ValueOf(&data).Elem()
				arrValue := reflect.ValueOf(out)
				kindCheck(arrValue)
				for i := 0; i < arrValue.Len(); i++ {
					dataValue.Set(reflect.Append(dataValue, arrValue.Index(i)))
				}
				p.tmpData = append(p.tmpData, data...)
			}
		},
	}
	t.evaluate(&ForEachOp{})
	t.data = p.tmpData
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t

}

func (p *pipeline) FindFirst(predicate Predicate) interface{} {
	nilCheck(predicate)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			if p.tmpData == nil {
				match := predicate(v)
				if match {
					p.tmpData = append(p.tmpData, v)
					p.sourceStage.stop = true
				}
			}
		},
	}
	t.evaluate(&ForEachOp{})
	if p.tmpData == nil {
		return nil
	}
	return p.tmpData[0]
}

func (p *pipeline) MaxMin(comparator Comparator) interface{} {
	nilCheck(comparator)
	return p.Reduce(func(t, u interface{}) interface{} {
		if comparator(t, u) {
			return t
		}
		return u
	})
}

func (p *pipeline) ToSlice(targetSlice interface{}) {
	nilCheck(targetSlice)
	targetValue := reflect.ValueOf(targetSlice)
	if targetValue.Kind() != reflect.Ptr {
		panic("target slice must be a pointer")
	}
	kindCheck(targetValue)
	sliceValue := reflect.Indirect(targetValue)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if v != nil {
				if p.sourceStage.parallel {
					p.lock.Lock()
					defer p.lock.Unlock()
				}
				sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(v)))
			}
		},
	}
	t.evaluate(&ForEachOp{})
}

func (p *pipeline) Reduce(function BiFunction) interface{} {
	nilCheck(function)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			if p.tmpData == nil {
				p.tmpData = append(p.tmpData, v)
			} else {
				res := function(p.tmpData[0], v)
				p.tmpData[0] = res
			}
		},
	}
	t.evaluate(&ForEachOp{})
	if p.tmpData == nil {
		return nil
	}
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
	entered, stop := p.matchOps(predicate, true)
	if entered {
		return stop
	}
	return false
}

func (p *pipeline) AllMatch(predicate Predicate) bool {
	entered, stop := p.matchOps(predicate, false)
	if entered {
		return !stop
	}
	return false
}

func (p *pipeline) matchOps(predicate Predicate, flag bool) (bool, bool) {
	nilCheck(predicate)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			p.sourceStage.entered = true
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
	return p.sourceStage.entered, p.sourceStage.stop
}

func (p *pipeline) Distinct(comparator Comparator) Stream {
	nilCheck(comparator)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			if p.tmpData == nil {
				p.tmpData = append(p.tmpData, v)
			} else {
				flag := true
				for _, tmp := range p.tmpData {
					if comparator(tmp, v) {
						flag = false
						break
					}
				}
				if flag {
					p.tmpData = append(p.tmpData, v)
				}
			}
		},
	}
	t.evaluate(&ForEachOp{})
	t.data = p.tmpData
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Sorted(comparator Comparator) Stream {
	nilCheck(comparator)
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
	nilCheck(consumer)
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
	nilCheck(predicate)
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
	nilCheck(consumer)
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
	nilCheck(function)
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			nextStage.do(nextStage.nextStage, function(v))
		},
	}
}

func (p *pipeline) evaluate(op TerminalOp) {
	nilCheck(op)
	for headStage := p; headStage != nil && headStage.previousStage != nil; headStage = headStage.previousStage {
		headStage.previousStage.nextStage = headStage
	}

	if p.sourceStage.parallel {
		op.EvaluateParallel(p.sourceStage)
	} else {
		op.EvaluateSequential(p.sourceStage)
	}
}

func (p *pipeline) statefulStage() *pipeline {
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			p.tmpData = append(p.tmpData, v)
		},
	}
}

func nilCheck(v interface{}) {
	if v == nil {
		panic("nil forbidden")
	}
}

func kindCheck(v reflect.Value) {
	nilCheck(v)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		panic("type must be Array or Slice")
	}
}
