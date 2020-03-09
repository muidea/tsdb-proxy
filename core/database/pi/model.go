package pi

// Result result
type Result struct {
	ErrorCode int    `json:"errorCode"`
	Result    string `json:"result"`
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
	TimeStamp string      `json:"timeStamp"`
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
