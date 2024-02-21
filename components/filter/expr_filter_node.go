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

package filter

//规则链节点配置示例：
//{
//        "id": "s2",
//        "type": "exprFilter",
//        "name": "表达式过滤器",
//        "debugMode": false,
//        "configuration": {
//          "expr": "msg.temperature > 50"
//        }
//      }
import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/rulego/rulego/api/types"
	"github.com/rulego/rulego/utils/json"
	"github.com/rulego/rulego/utils/maps"
)

func init() {
	Registry.Add(&ExprFilterNode{})
}

// ExprFilterNodeConfiguration 节点配置
type ExprFilterNodeConfiguration struct {
	// 表达式
	Expr string
}

// ExprFilterNode 使用expr表达式过滤消息
// 如果返回值`True`发送信息到`True`链, `False`发到`False`链。
// 如果表达式执行失败则发送到`Failure`链
// 通过`msg`变量访问消息体，如果消息的dataType是json类型，可以通过 `msg.XX`方式访问msg的字段。例如:`msg.temperature > 50;`
// 通过`metadata`变量访问消息元数据。例如 `metadata.customerName`
// 通过`msgType`变量访问消息类型
// 通过`dataType`变量访问数据类型
type ExprFilterNode struct {
	//节点配置
	Config  ExprFilterNodeConfiguration
	program *vm.Program
}

// Type 组件类型
func (x *ExprFilterNode) Type() string {
	return "exprFilter"
}
func (x *ExprFilterNode) New() types.Node {
	return &ExprFilterNode{Config: ExprFilterNodeConfiguration{
		Expr: "",
	}}
}

// Init 初始化
func (x *ExprFilterNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.Config)
	if err == nil {
		if program, err := expr.Compile(x.Config.Expr, expr.AllowUndefinedVariables(), expr.AsBool()); err == nil {
			x.program = program
		}
	}
	return err
}

// OnMsg 处理消息
func (x *ExprFilterNode) OnMsg(ctx types.RuleContext, msg types.RuleMsg) {

	var data interface{} = msg.Data
	if msg.DataType == types.JSON {
		var dataMap = make(map[string]interface{})
		if err := json.Unmarshal([]byte(msg.Data), &dataMap); err == nil {
			data = dataMap
		}
	}
	var evn = make(map[string]interface{})
	evn[types.MsgKey] = data
	evn[types.MetadataKey] = msg.Metadata
	evn[types.MsgTypeKey] = msg.Type
	evn[types.DataTypeKey] = msg.DataType

	if out, err := vm.Run(x.program, evn); err != nil {
		ctx.TellFailure(msg, err)
	} else {
		if result, ok := out.(bool); ok && result {
			ctx.TellNext(msg, types.True)
		} else {
			ctx.TellNext(msg, types.False)
		}
	}
}

// Destroy 销毁
func (x *ExprFilterNode) Destroy() {
}
