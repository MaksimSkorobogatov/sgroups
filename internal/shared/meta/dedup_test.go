package meta

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type dedupTestSuite struct {
	suite.Suite
}

func Test_Dedup(t *testing.T) {
	suite.Run(t, new(dedupTestSuite))
}

func (sui *dedupTestSuite) Test_Slices() {
	testCases := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "[]string",
			input:    []string{"a", "b", "a", "c", "b", "a"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "[]int",
			input:    []int{1, 2, 1, 3, 2, 1},
			expected: []int{1, 2, 3},
		},
		{
			name:     "[]uint32",
			input:    []uint32{10, 20, 10, 30, 20},
			expected: []uint32{10, 20, 30},
		},
		{
			name:     "[]int64",
			input:    []int64{100, 200, 100, 300},
			expected: []int64{100, 200, 300},
		},
		{
			name:     "[]bool",
			input:    []bool{true, false, true, false, true},
			expected: []bool{true, false},
		},
		{
			name:     "nil slice",
			input:    ([]string)(nil),
			expected: ([]string)(nil),
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "slice with single element",
			input:    []string{"single"},
			expected: []string{"single"},
		},
		{
			name:     "already unique slice",
			input:    []string{"unique1", "unique2", "unique3"},
			expected: []string{"unique1", "unique2", "unique3"},
		},
		{
			name:     "already unique int slice",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "all elements are same",
			input:    []string{"same", "same", "same", "same"},
			expected: []string{"same"},
		},
		{
			name:     "all ints are same",
			input:    []int{42, 42, 42, 42, 42},
			expected: []int{42},
		},
		{
			name:     "slice with empty strings",
			input:    []string{"", "a", "", "b", "a", ""},
			expected: []string{"", "a", "b"},
		},
		{
			name:     "preserve first occurrence order",
			input:    []string{"first", "second", "first", "third", "second", "fourth"},
			expected: []string{"first", "second", "third", "fourth"},
		},
		{
			name:     "preserve int order",
			input:    []int{5, 3, 5, 1, 3, 7},
			expected: []int{5, 3, 1, 7},
		},
	}

	for _, tc := range testCases {
		tc := tc
		sui.Run(tc.name, func() {
			DedupCollections(&tc.input)
			sui.Require().Equal(tc.expected, tc.input)
		})
	}
}

func (sui *dedupTestSuite) Test_StructsWithSliceFields() {
	type SimpleStruct struct {
		Names []string
		Ids   []int
	}

	sui.Run("struct with slice fields", func() {
		input := SimpleStruct{
			Names: []string{"a", "b", "a", "c", "b"},
			Ids:   []int{1, 2, 1, 3, 2},
		}
		expected := SimpleStruct{
			Names: []string{"a", "b", "c"},
			Ids:   []int{1, 2, 3},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("pointer to struct", func() {
		input := &SimpleStruct{
			Names: []string{"a", "b", "a", "c", "b"},
			Ids:   []int{1, 2, 1, 3, 2},
		}
		expected := &SimpleStruct{
			Names: []string{"a", "b", "c"},
			Ids:   []int{1, 2, 3},
		}
		DedupCollections(input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("pointer to pointer to struct", func() {
		input := &SimpleStruct{
			Names: []string{"a", "b", "a", "c", "b"},
			Ids:   []int{1, 2, 1, 3, 2},
		}
		expected := &SimpleStruct{
			Names: []string{"a", "b", "c"},
			Ids:   []int{1, 2, 3},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("struct with nil slice fields", func() {
		input := SimpleStruct{
			Names: nil,
			Ids:   nil,
		}
		DedupCollections(&input)
		sui.Require().Nil(input.Names)
		sui.Require().Nil(input.Ids)
	})

	sui.Run("struct with empty slice fields", func() {
		input := SimpleStruct{
			Names: []string{},
			Ids:   []int{},
		}
		DedupCollections(&input)
		sui.Require().Equal([]string{}, input.Names)
		sui.Require().Equal([]int{}, input.Ids)
	})
}

func (sui *dedupTestSuite) Test_NestedStructs() {
	type Inner struct {
		Values []string
	}
	type Outer struct {
		Inner Inner
		Tags  []string
	}

	sui.Run("nested struct with slices", func() {
		input := Outer{
			Inner: Inner{
				Values: []string{"x", "y", "x", "z"},
			},
			Tags: []string{"tag1", "tag2", "tag1"},
		}
		expected := Outer{
			Inner: Inner{
				Values: []string{"x", "y", "z"},
			},
			Tags: []string{"tag1", "tag2"},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	type OuterWithPtr struct {
		Inner *Inner
		Tags  []string
	}

	sui.Run("struct with pointer to nested struct", func() {
		input := OuterWithPtr{
			Inner: &Inner{
				Values: []string{"a", "b", "a", "c"},
			},
			Tags: []string{"t1", "t2", "t1"},
		}
		expected := OuterWithPtr{
			Inner: &Inner{
				Values: []string{"a", "b", "c"},
			},
			Tags: []string{"t1", "t2"},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("struct with nil pointer to nested struct", func() {
		input := OuterWithPtr{
			Inner: nil,
			Tags:  []string{"t1", "t2", "t1"},
		}
		DedupCollections(&input)
		sui.Require().Nil(input.Inner)
		sui.Require().Equal([]string{"t1", "t2"}, input.Tags)
	})
}

func (sui *dedupTestSuite) Test_DeeplyNestedStructs() {
	type Level3 struct {
		Items []int
	}
	type Level2 struct {
		L3   Level3
		Data []string
	}
	type Level1 struct {
		L2    Level2
		Names []string
	}

	sui.Run("deeply nested structs", func() {
		input := Level1{
			L2: Level2{
				L3: Level3{
					Items: []int{1, 2, 1, 3, 2},
				},
				Data: []string{"a", "b", "a"},
			},
			Names: []string{"x", "y", "x", "z"},
		}
		expected := Level1{
			L2: Level2{
				L3: Level3{
					Items: []int{1, 2, 3},
				},
				Data: []string{"a", "b"},
			},
			Names: []string{"x", "y", "z"},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})
}

func (sui *dedupTestSuite) Test_Maps() {
	sui.Run("map with slice values", func() {
		input := map[string][]string{
			"key1": {"a", "b", "a", "c"},
			"key2": {"x", "y", "x"},
		}
		expected := map[string][]string{
			"key1": {"a", "b", "c"},
			"key2": {"x", "y"},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("map with int slice values", func() {
		input := map[string][]int{
			"nums1": {1, 2, 1, 3},
			"nums2": {10, 20, 10},
		}
		expected := map[string][]int{
			"nums1": {1, 2, 3},
			"nums2": {10, 20},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("nil map", func() {
		var input map[string][]string
		DedupCollections(&input)
		sui.Require().Nil(input)
	})

	sui.Run("empty map", func() {
		input := map[string][]string{}
		DedupCollections(&input)
		sui.Require().Equal(map[string][]string{}, input)
	})
}

func (sui *dedupTestSuite) Test_MapsWithStructValues() {
	type ValueStruct struct {
		Items []string
		Nums  []int
	}

	sui.Run("map with struct values containing slices", func() {
		input := map[string]ValueStruct{
			"key1": {
				Items: []string{"a", "b", "a"},
				Nums:  []int{1, 2, 1},
			},
			"key2": {
				Items: []string{"x", "y", "x", "z"},
				Nums:  []int{10, 20, 10, 30},
			},
		}
		expected := map[string]ValueStruct{
			"key1": {
				Items: []string{"a", "b"},
				Nums:  []int{1, 2},
			},
			"key2": {
				Items: []string{"x", "y", "z"},
				Nums:  []int{10, 20, 30},
			},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("map with pointer to struct values", func() {
		input := map[string]*ValueStruct{
			"key1": {
				Items: []string{"a", "b", "a"},
				Nums:  []int{1, 2, 1},
			},
		}
		DedupCollections(&input)
		sui.Require().Equal([]string{"a", "b"}, input["key1"].Items)
		sui.Require().Equal([]int{1, 2}, input["key1"].Nums)
	})
}

func (sui *dedupTestSuite) Test_StructsWithMaps() {
	type StructWithMap struct {
		Data map[string][]string
		Tags []string
	}

	sui.Run("struct containing map with slice values", func() {
		input := StructWithMap{
			Data: map[string][]string{
				"key1": {"a", "b", "a"},
				"key2": {"x", "y", "x"},
			},
			Tags: []string{"t1", "t2", "t1"},
		}
		expected := StructWithMap{
			Data: map[string][]string{
				"key1": {"a", "b"},
				"key2": {"x", "y"},
			},
			Tags: []string{"t1", "t2"},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})
}

func (sui *dedupTestSuite) Test_NestedMaps() {
	sui.Run("map of maps with slice values", func() {
		input := map[string]map[string][]int{
			"outer1": {
				"inner1": {1, 2, 1, 3},
				"inner2": {10, 20, 10},
			},
			"outer2": {
				"inner3": {100, 200, 100},
			},
		}
		expected := map[string]map[string][]int{
			"outer1": {
				"inner1": {1, 2, 3},
				"inner2": {10, 20},
			},
			"outer2": {
				"inner3": {100, 200},
			},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})
}

func (sui *dedupTestSuite) Test_SlicesOfStructs() {
	type Item struct {
		Tags []string
		Ids  []int
	}

	sui.Run("slice of structs with slice fields", func() {
		input := []Item{
			{
				Tags: []string{"a", "b", "a"},
				Ids:  []int{1, 2, 1},
			},
			{
				Tags: []string{"x", "y", "x"},
				Ids:  []int{10, 20, 10},
			},
		}
		expected := []Item{
			{
				Tags: []string{"a", "b"},
				Ids:  []int{1, 2},
			},
			{
				Tags: []string{"x", "y"},
				Ids:  []int{10, 20},
			},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("slice of pointers to structs", func() {
		input := []*Item{
			{
				Tags: []string{"a", "b", "a"},
				Ids:  []int{1, 2, 1},
			},
			{
				Tags: []string{"x", "y", "x"},
				Ids:  []int{10, 20, 10},
			},
		}
		DedupCollections(&input)
		sui.Require().Equal([]string{"a", "b"}, input[0].Tags)
		sui.Require().Equal([]int{1, 2}, input[0].Ids)
		sui.Require().Equal([]string{"x", "y"}, input[1].Tags)
		sui.Require().Equal([]int{10, 20}, input[1].Ids)
	})
}

func (sui *dedupTestSuite) Test_Arrays() {
	sui.Run("array deduplication", func() {
		input := [5]int{1, 2, 1, 3, 2}
		expected := [5]int{1, 2, 3, 0, 0}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	sui.Run("string array deduplication", func() {
		input := [4]string{"a", "b", "a", "c"}
		expected := [4]string{"a", "b", "c", ""}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})

	type StructWithArray struct {
		Items [5]int
		Tags  []string
	}

	sui.Run("struct with array field", func() {
		input := StructWithArray{
			Items: [5]int{1, 2, 1, 3, 2},
			Tags:  []string{"a", "b", "a"},
		}
		expected := StructWithArray{
			Items: [5]int{1, 2, 3, 0, 0},
			Tags:  []string{"a", "b"},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})
}

func (sui *dedupTestSuite) Test_CircularReferences() {
	type Node struct {
		Value string
		Tags  []string
		Next  *Node
	}

	sui.Run("circular reference in linked list", func() {
		node1 := &Node{
			Value: "node1",
			Tags:  []string{"a", "b", "a"},
		}
		node2 := &Node{
			Value: "node2",
			Tags:  []string{"x", "y", "x"},
			Next:  node1,
		}
		node1.Next = node2 // circular reference

		DedupCollections(node1)

		// Both nodes should be deduplicated without infinite recursion
		sui.Require().Equal([]string{"a", "b"}, node1.Tags)
		sui.Require().Equal([]string{"x", "y"}, node2.Tags)
		sui.Require().Same(node2, node1.Next)
		sui.Require().Same(node1, node2.Next)
	})

	sui.Run("self-referencing struct", func() {
		node := &Node{
			Value: "self",
			Tags:  []string{"a", "b", "a", "c"},
		}
		node.Next = node // self-reference

		DedupCollections(node)

		sui.Require().Equal([]string{"a", "b", "c"}, node.Tags)
		sui.Require().Same(node, node.Next)
	})
}

func (sui *dedupTestSuite) Test_ComplexNestedStructure() {
	type Item struct {
		Values []int
	}
	type Container struct {
		Items    []Item
		Tags     []string
		Metadata map[string][]string
	}
	type Root struct {
		Containers []Container
		GlobalTags []string
	}

	sui.Run("complex nested structure", func() {
		input := Root{
			Containers: []Container{
				{
					Items: []Item{
						{Values: []int{1, 2, 1, 3}},
						{Values: []int{10, 20, 10}},
					},
					Tags: []string{"a", "b", "a"},
					Metadata: map[string][]string{
						"key1": {"x", "y", "x"},
					},
				},
				{
					Items: []Item{
						{Values: []int{100, 200, 100}},
					},
					Tags: []string{"c", "d", "c"},
					Metadata: map[string][]string{
						"key2": {"m", "n", "m"},
					},
				},
			},
			GlobalTags: []string{"global1", "global2", "global1"},
		}

		expected := Root{
			Containers: []Container{
				{
					Items: []Item{
						{Values: []int{1, 2, 3}},
						{Values: []int{10, 20}},
					},
					Tags: []string{"a", "b"},
					Metadata: map[string][]string{
						"key1": {"x", "y"},
					},
				},
				{
					Items: []Item{
						{Values: []int{100, 200}},
					},
					Tags: []string{"c", "d"},
					Metadata: map[string][]string{
						"key2": {"m", "n"},
					},
				},
			},
			GlobalTags: []string{"global1", "global2"},
		}

		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})
}

func (sui *dedupTestSuite) Test_MixedTypes() {
	type Mixed struct {
		Strings []string
		Ints    []int
		Uints   []uint32
		Bools   []bool
		Data    map[string][]int
	}

	sui.Run("struct with multiple slice types", func() {
		input := Mixed{
			Strings: []string{"a", "b", "a", "c"},
			Ints:    []int{1, 2, 1, 3, 2},
			Uints:   []uint32{10, 20, 10, 30},
			Bools:   []bool{true, false, true, false},
			Data: map[string][]int{
				"key": {100, 200, 100},
			},
		}
		expected := Mixed{
			Strings: []string{"a", "b", "c"},
			Ints:    []int{1, 2, 3},
			Uints:   []uint32{10, 20, 30},
			Bools:   []bool{true, false},
			Data: map[string][]int{
				"key": {100, 200},
			},
		}
		DedupCollections(&input)
		sui.Require().Equal(expected, input)
	})
}

func (sui *dedupTestSuite) Test_PassByValue() {
	sui.Run("no-op when passed by value (struct)", func() {
		type SimpleStruct struct {
			Items []string
		}
		original := SimpleStruct{
			Items: []string{"a", "b", "a"},
		}
		DedupCollections(original)
		sui.Require().Equal([]string{"a", "b", "a"}, original.Items)
	})

	sui.Run("no-op when passed by value (slice)", func() {
		original := []int{1, 2, 1, 3, 2}
		DedupCollections(original)
		sui.Require().Equal([]int{1, 2, 1, 3, 2}, original)
	})

	sui.Run("dedup applied when passed by pointer", func() {
		type SimpleStruct struct {
			Items []string
		}
		original := SimpleStruct{
			Items: []string{"a", "b", "a"},
		}
		DedupCollections(&original)
		sui.Require().Equal([]string{"a", "b"}, original.Items)
	})
}

func (sui *dedupTestSuite) Test_SharedSliceReferences() {
	sui.Run("multiple struct fields referencing same slice", func() {
		shared := []string{"a", "b", "a", "c"}
		type Shared struct {
			Field1 []string
			Field2 []string
		}
		input := Shared{
			Field1: shared,
			Field2: shared,
		}
		DedupCollections(&input)
		// Both fields reference the same slice, so dedup happens only once
		sui.Require().Equal([]string{"a", "b", "c"}, input.Field1)
		sui.Require().Equal([]string{"a", "b", "c"}, input.Field2)
		// They should still be the same slice (same pointer)
		sui.Require().True(len(input.Field1) == len(input.Field2))
	})
}
