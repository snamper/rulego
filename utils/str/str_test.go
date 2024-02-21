/*
 * Copyright 2023 The RuleGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package str

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rulego/rulego/test/assert"
	"math"
	"reflect"
	"testing"
)

func TestProcessVar(t *testing.T) {
	s := ProcessVar("Hello, Alice. You are ${age} years old.", "age", "18")
	assert.Equal(t, "Hello, Alice. You are 18 years old.", s)
}
func TestProcessVar2(t *testing.T) {
	var input interface{}
	input = float64(5)
	newValue, _ := json.Marshal(input)
	fmt.Println(string(newValue))
	fmt.Println(ToString(input))
	input = true
	newValue, _ = json.Marshal(input)
	fmt.Println(string(newValue))
	fmt.Println(ToString(input))
	input = "aa"
	newValue, _ = json.Marshal(input)
	fmt.Println(string(newValue))
	fmt.Println(ToString(input))
	input = []byte("bb")
	newValue, _ = json.Marshal(input)
	fmt.Println(string(newValue))
	fmt.Println(ToString(input))

}

func TestSprintfDict(t *testing.T) {
	// 创建一个字典
	dict := map[string]string{
		"name": "Alice",
		"age":  "18",
	}
	// 使用SprintfDict来格式化字符串
	s := SprintfDict("Hello, ${name}. You are ${age} years old.", dict)
	assert.Equal(t, "Hello, Alice. You are 18 years old.", s)
}

func TestSprintfVar(t *testing.T) {
	// 创建一个字典
	dict := map[string]string{
		"name": "Alice",
		"age":  "18",
	}
	// 使用SprintfDict来格式化字符串
	s := SprintfVar("Hello, ${global.name}. You are ${global.age} years old.", "global.", dict)
	assert.Equal(t, "Hello, Alice. You are 18 years old.", s)
}

type Stringer struct {
	Value string
}

func (s *Stringer) String() string {
	return s.Value
}

func TestToString(t *testing.T) {
	// Test cases
	testCases := []struct {
		want  string
		input interface{}
	}{
		{"123", int(123)},
		{"123", uint(123)},
		{"123", int8(123)},
		{"123", uint8(123)},
		{"123", int16(123)},
		{"123", uint16(123)},
		{"123", int32(123)},
		{"123", uint32(123)},
		{"123", int64(123)},
		{"123", uint64(123)},
		{"3.14", float32(3.14)},
		{"3.14", float64(3.14)},
		{"true", true},
		{"hello", &Stringer{"hello"}},
		{"hello", []byte("hello")},
		{"", nil},
		{"", ""},
		{"hello", "hello"},
		{"error", errors.New("error")},
		{"{\"Username\":\"lala\",\"Age\":25,\"Address\":{\"Detail\":\"\"}}", User{Username: "lala", Age: 25}},
		{"{\"name\":\"lala\"}", map[string]string{
			"name": "lala",
		}},
	}

	for _, tc := range testCases {
		s := ToString(tc.input)
		assert.Equal(t, tc.want, s)
	}
}

func TestToStringMaybeErr(t *testing.T) {
	// Test cases
	testCases := []struct {
		want  string
		input interface{}
	}{
		{"123", int(123)},
		{"123", uint(123)},
		{"123", int8(123)},
		{"123", uint8(123)},
		{"123", int16(123)},
		{"123", uint16(123)},
		{"123", int32(123)},
		{"123", uint32(123)},
		{"123", int64(123)},
		{"123", uint64(123)},
		{"3.14", float32(3.14)},
		{"3.14", float64(3.14)},
		{"true", true},
		{"hello", &Stringer{"hello"}},
		{"hello", []byte("hello")},
		{"", nil},
		{"", ""},
		{"hello", "hello"},
		{"error", errors.New("error")},
		{"{\"Username\":\"lala\",\"Age\":25,\"Address\":{\"Detail\":\"\"}}", User{Username: "lala", Age: 25}},
		{"{\"name\":\"lala\"}", map[string]string{
			"name": "lala",
		}},
		{"json: unsupported value: NaN", map[string]interface{}{
			"name": math.Sqrt(-1),
		}},
	}

	for _, tc := range testCases {
		s, err := ToStringMaybeErr(tc.input)
		if err != nil {
			assert.Equal(t, tc.want, err.Error())
		} else {
			assert.Equal(t, tc.want, s)
		}
	}
}
func TestToStringMapString(t *testing.T) {
	// Test cases
	testCases := []struct {
		input interface{}
	}{
		{map[string]interface{}{
			"name": "lala",
			"age":  5,
			"user": User{},
		}},
		{map[interface{}]string{
			"name": "lala",
		}},
		{map[string]string{
			"name": "lala",
		}},
		{map[interface{}]interface{}{
			"name": "lala",
		}},
		{"{\"name\":\"lala\"}"},
	}

	for _, tc := range testCases {
		strMap := ToStringMapString(tc.input)
		nameV := strMap["name"]
		nameType := reflect.TypeOf(&nameV).Elem()
		assert.Equal(t, reflect.String, nameType.Kind())
		assert.Equal(t, "lala", nameV)
	}

	strMap := ToStringMapString(&User{})
	assert.Equal(t, 0, len(strMap))
}

func TestRandomStr(t *testing.T) {
	v1 := RandomStr(10)
	assert.Equal(t, 10, len(v1))
	v2 := RandomStr(10)
	assert.Equal(t, 10, len(v2))
	v3 := RandomStr(4)
	assert.Equal(t, 4, len(v3))
	assert.True(t, v1 != v2)
}

func TestCheckHasVar(t *testing.T) {
	assert.True(t, CheckHasVar("${ddd}"))
	assert.False(t, CheckHasVar("${ddd"))
	assert.True(t, CheckHasVar("${ ddd }"))
	assert.False(t, CheckHasVar("ddd"))
}

func TestConvertDollarPlaceholder(t *testing.T) {
	sql := "select * from user where name=? and age=?"
	assert.Equal(t, "select * from user where name=$1 and age=$2", ConvertDollarPlaceholder(sql, "postgres"))
	assert.Equal(t, sql, ConvertDollarPlaceholder(sql, "mysql"))
}

func TestToLowerFirst(t *testing.T) {
	assert.Equal(t, "", ToLowerFirst(""))
	assert.Equal(t, "hello", ToLowerFirst("Hello"))
	assert.Equal(t, "hello", ToLowerFirst("hello"))
	assert.Equal(t, "hELLO", ToLowerFirst("HELLO"))
}

func TestRemoveBraces(t *testing.T) {
	assert.Equal(t, "", RemoveBraces(""))
	assert.Equal(t, "hello", RemoveBraces("hello"))
	assert.Equal(t, "hello_lala", RemoveBraces("${hello_lala}"))
	assert.Equal(t, "hello", RemoveBraces("hello}"))
	assert.Equal(t, "hello", RemoveBraces("${hello}"))
	assert.Equal(t, "helloage", RemoveBraces("${hello} ${age}"))
}

type User struct {
	Username string
	Age      int
	Address  Address
}
type Address struct {
	Detail string
}
