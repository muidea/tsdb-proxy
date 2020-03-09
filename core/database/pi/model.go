package pi

import (
	pb "supos.ai/data-lake/external/tsdb-proxy/common/model"
)

const (
	typeNull      = 0
	typeBool      = 1
	typeUint8     = 2
	typeInt8      = 3
	typeChar      = 4
	typeUint16    = 5
	typeInt16     = 6
	typeUint32    = 7
	typeInt32     = 8
	typeUint64    = 9
	typeInt64     = 10
	typeFloat16   = 11
	typeFloat32   = 12
	typeFloat64   = 13
	typeDigital   = 101
	typeBlob      = 102
	typeTimestamp = 104
	typeString    = 105
	typeBad       = 255
)

// Result result
type Result struct {
	ErrorCode int    `json:"errorCode"`
	Reason    string `json:"reason"`
}

// TagInfo tag info
type TagInfo struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
	Type int    `json:"type"`
	Unit string `json:"unit"`
}

// TagInfoList tagInfo list
type TagInfoList []*TagInfo

// TagValue tag value
type TagValue struct {
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Quality   int         `json:"quality"`
	TimeStamp int64       `json:"timeStamp"`
}

// TagValueList tagValue list
type TagValueList []*TagValue

// EnumTagsData enum tag data
type EnumTagsData struct {
	Tags TagInfoList `json:"tags"`
}

// EnumTagsResult enum tag result
type EnumTagsResult struct {
	Result
	Data EnumTagsData `json:"data"`
}

// SubscribeParam subscribe param
type SubscribeParam struct {
	Tags     []string `json:"tags"`
	CallBack string   `json:"callback"`
}

// SubscribeResult subscribe result
type SubscribeResult Result

// UnsubscribeParam unsubscribe param
type UnsubscribeParam SubscribeParam

// UnsubscribeResult unsubscirbe result
type UnsubscribeResult SubscribeResult

// NotifyValuesData notify values data
type NotifyValuesData struct {
	TimeStamp string       `json:"timeStamp"`
	Value     TagValueList `json:"value"`
}

// QueryParam query param
type QueryParam struct {
	BeginTime string   `json:"beginTime"`
	EndTime   string   `json:"endTime"`
	Count     int      `json:"count"`
	Tags      []string `json:"tags"`
}

// QueryData query data
type QueryData struct {
	Values map[string]TagValueList `json:"values"`
}

// QueryResult query result
type QueryResult struct {
	Result
	Data QueryData `json:"data"`
}

// CheckHealthData check health data
type CheckHealthData struct {
	Status int `json:"status"`
}

// CheckHealthResult check health result
type CheckHealthResult struct {
	Result
	Data CheckHealthData `json:"data"`
}

func getPrimiteType(typeVal int) string {
	switch typeVal {
	case typeBool:
		return "Boolean"
	case typeUint8,
		typeInt8,
		typeUint16,
		typeInt16,
		typeInt32:
		return "Integer"
	case typeUint32,
		typeUint64,
		typeInt64:
		return "Long"
	case typeFloat16,
		typeFloat32:
		return "Float"
	case typeFloat64:
		return "Double"
	case typeString:
		return "String"
	default:
	}

	return ""
}

// TagInfo2Property tagInfo to MetaProperty
func TagInfo2Property(info *TagInfo, property *pb.MetaProperty) bool {
	property.Name = info.Name
	property.Description = info.Desc
	property.PrimitiveType = getPrimiteType(info.Type)
	return property.PrimitiveType != ""
}

// TagValue2NameValue tagValue to NamedValue
func TagValue2NameValue(val *TagValue, info *TagInfo, nv *pb.NamedValue) bool {
	pv := &pb.PrimitiveValue{}
	switch info.Type {
	case typeBool:
		pv.Value = &pb.PrimitiveValue_BoolValue{BoolValue: val.Value.(bool)}
	case typeUint8,
		typeInt8,
		typeUint16,
		typeInt16,
		typeInt32:
		pv.Value = &pb.PrimitiveValue_I32Value{I32Value: val.Value.(int32)}
	case typeUint32:
		pv.Value = &pb.PrimitiveValue_Ui32Value{Ui32Value: val.Value.(uint32)}
	case typeUint64:
		pv.Value = &pb.PrimitiveValue_Ui64Value{Ui64Value: val.Value.(uint64)}
	case typeInt64:
		pv.Value = &pb.PrimitiveValue_I64Value{I64Value: val.Value.(int64)}
	case typeFloat16,
		typeFloat32:
		pv.Value = &pb.PrimitiveValue_FltValue{FltValue: val.Value.(float32)}
	case typeFloat64:
		pv.Value = &pb.PrimitiveValue_DblValue{DblValue: val.Value.(float64)}
	case typeString:
		pv.Value = &pb.PrimitiveValue_StrValue{StrValue: val.Value.(string)}
	default:
	}

	nv.Name = val.Name
	nv.Value = &pb.Value{
		Kind: &pb.Value_PrimitiveValueWithQT{
			PrimitiveValueWithQT: &pb.PrimitiveValueWithQT{
				Time:    uint64(val.TimeStamp) * 1000,
				Quality: uint64(val.Quality),
				Status:  0,
				Value:   pv,
			}}}

	return true
}
