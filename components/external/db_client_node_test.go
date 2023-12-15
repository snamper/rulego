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

package external

import (
	"fmt"
	"github.com/rulego/rulego/api/types"
	"github.com/rulego/rulego/test"
	"github.com/rulego/rulego/test/assert"
	"github.com/rulego/rulego/utils/json"
	"github.com/rulego/rulego/utils/str"
	"testing"
	"time"
)

func TestDbClientNode(t *testing.T) {
	var targetNodeType = "dbClient"

	t.Run("NewNode", func(t *testing.T) {
		test.NodeNew(t, targetNodeType, &DbClientNode{}, types.Configuration{
			"sql":        "select * from test",
			"driverName": "mysql",
			"dsn":        "root:root@tcp(127.0.0.1:3306)/test",
		}, Registry)
	})

	//t.Run("InitNode", func(t *testing.T) {
	//	test.NodeInit(t, targetNodeType, types.Configuration{
	//		"sql":        "select * from test2",
	//		"driverName": "postgres",
	//		"dsn":        "root:root@tcp(127.0.0.1:3306)/test2",
	//	}, types.Configuration{
	//		"sql":        "select * from test2",
	//		"driverName": "postgres",
	//		"dsn":        "root:root@tcp(127.0.0.1:3306)/test2",
	//	}, Registry)
	//})
	//
	//t.Run("DefaultConfig", func(t *testing.T) {
	//	test.NodeInit(t, targetNodeType, types.Configuration{
	//		"sql":        "select * from test",
	//		"driverName": "mysql",
	//		"dsn":        "root:root@tcp(127.0.0.1:3306)/test",
	//	}, types.Configuration{
	//		"sql":        "select * from test",
	//		"driverName": "mysql",
	//		"dsn":        "root:root@tcp(127.0.0.1:3306)/test",
	//	}, Registry)
	//})

	t.Run("OnMsg", func(t *testing.T) {

		node, err := test.CreateAndInitNode(targetNodeType, types.Configuration{
			"sql":        "select * from test",
			"driverName": "mysql",
			"dsn":        "root:root@tcp(127.0.0.1:3308)/test",
		}, Registry)
		assert.NotNil(t, err)

		metaData := types.BuildMetadata(make(map[string]string))
		metaData.PutValue("productType", "test")
		msgList := []test.Msg{
			{
				MetaData:   metaData,
				MsgType:    "ACTIVITY_EVENT1",
				Data:       "AA",
				AfterSleep: time.Millisecond * 200,
			},
			{
				MetaData:   metaData,
				MsgType:    "ACTIVITY_EVENT2",
				Data:       "{\"temperature\":60}",
				AfterSleep: time.Millisecond * 200,
			},
		}

		var nodeList = []test.NodeAndCallback{
			{
				Node:    node,
				MsgList: msgList,
				Callback: func(msg types.RuleMsg, relationType string, err error) {
					assert.Equal(t, types.Failure, relationType)
				},
			},
		}
		for _, item := range nodeList {
			test.NodeOnMsgWithChildren(t, item.Node, item.MsgList, item.ChildrenNodes, item.Callback)
		}
	})
}

// 测试mysql增删修改查
func TestMysqlDbClientNodeOnMsg(t *testing.T) {
	testDbClientNodeOnMsg(t, "mysql", "root:root@tcp(127.0.0.1:3306)/test")
}

// 测试postgres增删修改查
func TestPgDbClientNodeOnMsg(t *testing.T) {
	testDbClientNodeOnMsg(t, "postgres", "postgres://postgres:postgres@127.0.0.1:5432/test?sslmode=disable")
}

func testDbClientNodeOnMsg(t *testing.T, driverName, dsn string) {

	metaData := types.NewMetadata()
	config := types.NewConfig()

	var configuration = make(types.Configuration)
	// 测试插入数据的操作
	configuration["sql"] = "insert into users (id,name, age) values (?,?,?)"
	configuration["params"] = []interface{}{"${id}", "${name}", "${age}"}
	configuration["poolSize"] = 10
	configuration["driverName"] = driverName
	configuration["dsn"] = dsn

	node := new(DbClientNode)

	err := node.Init(config, configuration)
	if err != nil {
		return
	}

	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err2 error) {
		assert.Equal(t, types.Success, relationType)
		assert.Equal(t, "1", msg.Metadata.GetValue(rowsAffectedKey))
	})
	metaData.PutValue("id", "1")
	metaData.PutValue("name", "test01")
	metaData.PutValue("age", "18")
	msg := ctx.NewMsg("TEST_MSG_TYPE_AA", metaData, "")
	node.OnMsg(ctx, msg)
	time.Sleep(time.Second)
	//插入第二条
	ctx = test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err2 error) {
		assert.Equal(t, types.Success, relationType)
		assert.Equal(t, "1", msg.Metadata.GetValue(rowsAffectedKey))
	})
	metaData.PutValue("id", "2")
	metaData.PutValue("name", "test02")
	metaData.PutValue("age", "35")
	msg = ctx.NewMsg("TEST_MSG_TYPE_BB", metaData, "")
	node.OnMsg(ctx, msg)

	// 测试查询一条记录的操作
	configuration["sql"] = "select * from users where id = ?"
	configuration["params"] = []interface{}{"${id}"}
	configuration["getOne"] = true
	configuration["poolSize"] = 10
	configuration["driverName"] = driverName
	configuration["dsn"] = dsn

	node = new(DbClientNode)
	err = node.Init(config, configuration)
	if err != nil {
		t.Errorf("err=%s", err)
	}

	ctx = test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err2 error) {
		// 检查查询结果是否正确
		assert.Equal(t, types.Success, relationType)
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(msg.Data), &result)
		assert.Equal(t, "1", str.ToString(result["id"]))
		assert.Equal(t, "test01", result["name"])
		fmt.Println(msg.Data)
	})

	metaData.PutValue("id", "1")
	msg = ctx.NewMsg("TEST_MSG_TYPE_CC", metaData, "")
	node.OnMsg(ctx, msg)

	// 测试查询多条记录的操作
	//不使用占位符参数
	configuration["sql"] = "select * from users where age >= ${age}"
	configuration["params"] = nil
	configuration["getOne"] = false
	configuration["poolSize"] = 10
	configuration["driverName"] = driverName
	configuration["dsn"] = dsn

	node = new(DbClientNode)
	err = node.Init(config, configuration)
	if err != nil {
		t.Errorf("err=%s", err)
	}

	ctx = test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err2 error) {
		// 检查查询结果是否正确
		assert.Equal(t, types.Success, relationType)
		var result []map[string]interface{}
		_ = json.Unmarshal([]byte(msg.Data), &result)
		assert.Equal(t, 2, len(result))
		fmt.Println(msg.Data)
	})

	metaData.PutValue("age", "10")
	msg = ctx.NewMsg("TEST_MSG_TYPE_DD", metaData, "")
	node.OnMsg(ctx, msg)

	// 测试修改数据的操作
	configuration["sql"] = "update users set age = ? where id = ?"
	configuration["params"] = []interface{}{"${age}", "${id}"}
	configuration["poolSize"] = 10
	configuration["driverName"] = driverName
	configuration["dsn"] = dsn

	node = new(DbClientNode)
	err = node.Init(config, configuration)
	if err != nil {
		t.Errorf("err=%s", err)
	}

	ctx = test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err2 error) {
		// 检查查询结果是否正确
		assert.Equal(t, types.Success, relationType)
		assert.Equal(t, "1", msg.Metadata.GetValue(rowsAffectedKey))
	})

	metaData.PutValue("id", "1")
	metaData.PutValue("age", "21")

	msg = ctx.NewMsg("TEST_MSG_TYPE_EE", metaData, "")
	node.OnMsg(ctx, msg)

	// 测试删除数据的操作
	configuration["sql"] = "delete from users"
	configuration["params"] = nil
	configuration["poolSize"] = 10
	configuration["driverName"] = driverName
	configuration["dsn"] = dsn

	node = new(DbClientNode)
	err = node.Init(config, configuration)
	if err != nil {
		t.Errorf("err=%s", err)
	}

	ctx = test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err2 error) {
		// 检查查询结果是否正确
		assert.Equal(t, types.Success, relationType)
		assert.Equal(t, "2", msg.Metadata.GetValue(rowsAffectedKey))
	})

	metaData.PutValue("id", "1")
	metaData.PutValue("age", "21")

	msg = ctx.NewMsg("TEST_MSG_TYPE_EE", metaData, "")
	node.OnMsg(ctx, msg)

	time.Sleep(time.Second * 2)
}
