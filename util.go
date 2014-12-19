package goformkeeper

import "sort"

type UniqueStringArrayBuilder struct {
	data map[string]bool
}

func NewUniqueStringArrayBuilder(capacity int) *UniqueStringArrayBuilder {
	return &UniqueStringArrayBuilder{
		data: make(map[string]bool, capacity),
	}
}

func (b *UniqueStringArrayBuilder) Add(value string) {
	b.data[value] = true
}

func (b *UniqueStringArrayBuilder) Build() []string {
	values := make([]string, 0)
	for value, _ := range b.data {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func Uniq(origin []string) []string {
	builder := NewUniqueStringArrayBuilder(len(origin))
	for _, v := range origin {
		builder.Add(v)
	}
	return builder.Build()
}
