// Scry Info.  All rights reserved.
// license that can be found in the license file.
package main

import (
	"encoding/json"
	"github.com/scryinfo/dot/dot"
	certificate "github.com/scryinfo/dot/dots/certificate"
	gindot "github.com/scryinfo/dot/dots/gindot"
	conns "github.com/scryinfo/dot/dots/grpc/conns"
	gserver "github.com/scryinfo/dot/dots/grpc/gserver"
	nobl_2 "github.com/scryinfo/dot/sample/grpc/nobl"
	nobl "github.com/scryinfo/dot/tools/config/data/nobl"
	"log"
	"os"
	"reflect"
)

//TypeLives living
type DotAndExtendConfig struct {
	Meta         dot.Metadata           `json:"metaData"`
	Lives        []Live                 `json:"lives"`
	RequiredInfo map[string]interface{} `json:"requiredInfo"`
}
type Live struct {
	TypeId    dot.TypeId            `json:"typeId"`
	LiveId    dot.LiveId            `json:"liveId"`
	RelyLives map[string]dot.LiveId `json:"relyLives"`
	Dot       dot.Dot
	Config    interface{} `json:"json"` //扩展配置
	Name      string      `json:"name"` //实例别名

}

func main() {
	//获取通用组件信息
	var result = make([]*dot.TypeLives, 0)
	{

		result = append(result, certificate.TypeLiveEcdsa())

		result = append(result, gindot.TypeLiveGinDot())
		result = append(result, gindot.TypeLiveRouter()...)

		result = append(result, conns.ConnNameTypeLives()...)
		result = append(result, conns.ConnsTypeLives())

		result = append(result, gserver.GinNoblTypeLives()...)
		result = append(result, gserver.HttpNoblTypeLives()...)
		result = append(result, gserver.ServerNoblTypeLive())

		result = append(result, nobl_2.HiClientTypeLives()...)
		result = append(result, nobl_2.HiServerTypeLives()...)

		result = append(result, nobl.RpcImplementTypeLives()...)
	}

	//初始化lives
	for i := range result {
		if result[i].Lives == nil {
			slice := make([]dot.Live, 0)
			slice = append(slice, dot.Live{})
			result[i].Lives = slice
		}
	}
	//对于typeId相同的组件进行合并
	var resultMerge = make([]*dot.TypeLives, 0)
	{
		//保存已经合并的组件
		merge := make(map[dot.TypeId]byte)
		leni := len(result)
		for i := 0; i < leni; i++ {
			//判断这个组件是否已经合并完毕
			_, ok := merge[result[i].Meta.TypeId]
			if ok {
				//跳过
			} else {
				for j := i + 1; j < leni; j++ {
					//判断是否具备合并条件
					if result[i].Meta.TypeId.String() == result[j].Meta.TypeId.String() {
						//合并
						//Meta部分
						result[i].Meta.Merge(&result[j].Meta)
						//lives部分
						if len(result[i].Lives[0].TypeId) > 0 {
							result[i].Lives[0].TypeId = result[j].Lives[0].TypeId
						}
						if len(result[i].Lives[0].LiveId) > 0 {
							result[i].Lives[0].LiveId = result[j].Lives[0].LiveId
						}
						for k, v := range result[j].Lives[0].RelyLives {
							if _, ok := result[i].Lives[0].RelyLives[k]; !ok {
								result[i].Lives[0].RelyLives[k] = v
							}
						}
					}
				}
				//将这个id放入merge中
				merge[result[i].Meta.TypeId] = 1
				resultMerge = append(resultMerge, result[i])
			}
		}
	}
	//获取组件特有的配置信息
	var configInfo = make([]*dot.ConfigTypeLives, 0)
	{

		configInfo = append(configInfo, gindot.ConfigTypeLiveGinDot())
		configInfo = append(configInfo, gindot.ConfigTypeLiveRouter())

		configInfo = append(configInfo, conns.ConnNameConfigTypeLives())
		configInfo = append(configInfo, conns.ConnsConfigTypeLives())

		configInfo = append(configInfo, gserver.HttpNoblConfigTypeLives())
		configInfo = append(configInfo, gserver.ServerNoblConfigTypeLive())

	}
	var finalResult = make([]*DotAndExtendConfig, 0)

	//将扩展配置以及组件信息加入最终结果中
	for i := range resultMerge {
		finalResult = append(finalResult, &DotAndExtendConfig{})
		//组件信息
		{
			//Meta
			finalResult[i].Meta = resultMerge[i].Meta
			//Lives
			finalResult[i].Lives = make([]Live, len(resultMerge[i].Lives))
			for key, value := range resultMerge[i].Lives {
				finalResult[i].Lives[key].TypeId = value.TypeId
				finalResult[i].Lives[key].LiveId = value.LiveId
				finalResult[i].Lives[key].RelyLives = value.RelyLives
				finalResult[i].Lives[key].Dot = value.Dot
			}
		}
		//扩展配置
		for j := range configInfo {
			if finalResult[i].Meta.TypeId.String() == configInfo[j].TypeIdConfig.String() {
				finalResult[i].Lives[0].Config = configInfo[j].ConfigInfo
			}
		}
	}
	//获取必填信息
	for key, _ := range finalResult {
		if finalResult[key].Lives[0].Config != nil {
			finalResult[key].RequiredInfo = getTags(reflect.TypeOf(finalResult[key].Lives[0].Config).Elem())
			//fmt.Println(reflect.TypeOf(finalResult[key].Lives[0].Config).Elem())
		}
	}
	//生成json文件
	{
		_, err := json.Marshal(finalResult)
		if err != nil {
			log.Fatal("MarShal err:", err)
		}
		file, _ := os.OpenFile("./run_out/result.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer file.Close()
		enc := json.NewEncoder(file)
		err = enc.Encode(finalResult)
		if err != nil {
			log.Println("Error in encoding json")
		}
	}
}

func getTags(t reflect.Type) map[string]interface{} {

	num := t.NumField()
	var tagResult = make(map[string]interface{})

	for i := 0; i < num; i++ {
		//todo 结构体应该做什么 ok
		ixField := t.Field(i)
		if ixField.Type.Kind() == reflect.Struct {
			if ixField.Tag.Get("json") == "" {
				tagResult[ixField.Name] = getTags(ixField.Type)
			} else {
				tagResult[ixField.Tag.Get("json")] = getTags(ixField.Type)
			}
			continue
		}
		//todo 指针应该做什么 ok
		if ixField.Type.Kind() == reflect.Ptr {
			//getTags(val.Field(i))
			//value:=reflect.Indirect(t.Field(i))
			//fmt.Println(t.Field(i).Type.Elem())
			if ixField.Type.Elem().Kind() == reflect.Struct {
				if ixField.Tag.Get("json") == "" {
					tagResult[ixField.Name] = getTags(ixField.Type.Elem())
				} else {
					tagResult[ixField.Tag.Get("json")] = getTags(ixField.Type.Elem())
				}
				continue
			}

		}
		//todo 切片应该做什么 ok
		if ixField.Type.Kind() == reflect.Slice {
			if ixField.Type.Elem().Kind() == reflect.Struct {
				value := reflect.MakeSlice(ixField.Type, 1, 1)
				//fmt.Println(value.Index(0))
				if ixField.Tag.Get("json") == "" {
					tagResult[ixField.Name] = getTags(value.Index(0).Type())
				} else {
					tagResult[ixField.Tag.Get("json")] = getTags(value.Index(0).Type())
				}
				continue
			}
		}
		//todo 数组应该做什么　ok
		if ixField.Type.Kind() == reflect.Array {
			if ixField.Type.Elem().Kind() == reflect.Struct {
				//fmt.Println(t.Field(i).Type.Kind())
				if ixField.Tag.Get("json") == "" {
					tagResult[ixField.Name] = getTags(ixField.Type.Elem())
				} else {
					tagResult[ixField.Tag.Get("json")] = getTags(ixField.Type.Elem())
				}
				continue
			}
		}
		//todo 映射应该做什么
		if ixField.Type.Kind() == reflect.Map {
			if ixField.Type.Elem().Kind() == reflect.Struct {
				if ixField.Tag.Get("json") == "" {
					tagResult[ixField.Name] = getTags(ixField.Type.Elem())
				} else {
					tagResult[ixField.Tag.Get("json")] = getTags(ixField.Type.Elem())
				}
				continue
			}
			//fmt.Println(t.Field(i).Type.Key())  //k
			//fmt.Println(t.Field(i).Type.Elem()) //v
			//getTags(t.Field(i).Type.Elem()) //?行不通　当ｖ的类型是简单类型时panic
			/*value:=reflect.MakeMap(t.Field(i).Type)
			tagResult[t.Field(i).Name]=getTags(value.Type().Elem())*/
			//continue
		}
		//todo 接口应该做什么
		if ixField.Type.Kind() == reflect.Interface {
			//在扩展配置内部不会有interface类型的字段，因为无法完成反序列化
			//使用value能实现类型的判断
			//func ValueOf(i interface{}) Value
			continue
		}

		if ixField.Tag.Get("required") == "yes" {
			//fmt.Println(ixField.Name + "字段是必填字段")
			if ixField.Tag.Get("json") == "" {
				tagResult[ixField.Name] = true
			} else {
				tagResult[ixField.Tag.Get("json")] = true
			}
		} else {
			//fmt.Println(ixField.Name + "字段不是必填字段")
			if ixField.Tag.Get("json") == "" {
				tagResult[ixField.Name] = false
			} else {
				tagResult[ixField.Tag.Get("json")] = false
			}
		}

	}
	return tagResult

}
