package pi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/muidea/magicCommon/foundation/net"
	"supos.ai/data-lake/external/tsdb-proxy/collector"
	pb "supos.ai/data-lake/external/tsdb-proxy/common/model"
	"supos.ai/data-lake/external/tsdb-proxy/core/database"
	"supos.ai/data-lake/external/tsdb-proxy/model"
)

type piImpl struct {
	info              *model.DBInfo
	subscribeCallBack string

	httpClient *http.Client
	collector  collector.Collector

	tagsInfo TagInfoList
}

// NewPi new pi DB
func NewPi(info *model.DBInfo, callBack string) database.DB {
	return &piImpl{info: info, subscribeCallBack: callBack, httpClient: &http.Client{}}
}

func (s *piImpl) Initialize(rtdService string) (err error) {
	s.collector = collector.NewCollector(s.info.Name)

	err = s.collector.Start(rtdService, s.OnStatusCallBack)

	return
}

func (s *piImpl) Uninitialize() {
	s.collector.Stop()
}

func (s *piImpl) QueryHistory(res http.ResponseWriter, req *http.Request) (err error) {
	return
}

func (s *piImpl) UpdateValue(res http.ResponseWriter, req *http.Request) (err error) {
	values := &pb.ValueSequnce{}

	err = s.collector.UpdateValue(values)
	return
}

func (s *piImpl) CheckHealth() {

}

func (s *piImpl) OnStatusCallBack(collectName string, status, errorCode int, reason string) {
	if status == collector.LoginStatus {
		if s.collector.IsReady() {
			values := &pb.ValueSequnce{}
			s.collector.UpdateValue(values)
		}
	}
}

func (s *piImpl) enumTags() (ret TagInfoList, err error) {
	url, _ := url.ParseRequestURI(s.info.Address)
	url.Path = strings.Join([]string{url.Path, "/tags/enum"}, "")

	result := &EnumTagsResult{}
	_, err = net.HTTPGet(s.httpClient, url.String(), result)
	if err != nil {
		return
	}

	if result.ErrorCode != 0 {
		err = fmt.Errorf("enum tags failed, erro:%s", result.Reason)
		return
	}

	ret = result.Data.Tags
	return
}

func (s *piImpl) subscribe(tags []string, callBack string) (err error) {
	url, _ := url.ParseRequestURI(s.info.Address)
	url.Path = strings.Join([]string{url.Path, "/realtime/subscribe"}, "")

	param := &SubscribeParam{Tags: tags, CallBack: callBack}
	result := &SubscribeResult{}
	_, err = net.HTTPPost(s.httpClient, url.String(), param, result)
	if err != nil {
		return
	}

	if result.ErrorCode != 0 {
		err = fmt.Errorf("subscribe failed, erro:%s", result.Reason)
	}

	return
}

func (s *piImpl) unsubscribe(tags []string, callBack string) (err error) {
	url, _ := url.ParseRequestURI(s.info.Address)
	url.Path = strings.Join([]string{url.Path, "/realtime/unsubscribe"}, "")

	param := &UnsubscribeParam{Tags: tags, CallBack: callBack}
	result := &UnsubscribeResult{}
	_, err = net.HTTPPost(s.httpClient, url.String(), param, result)
	if err != nil {
		return
	}

	if result.ErrorCode != 0 {
		err = fmt.Errorf("unsubscribe failed, erro:%s", result.Reason)
	}

	return
}

func (s *piImpl) queryHistory(beginTime, endTime string, valueCount int, tags []string) (ret map[string]TagValueList, err error) {
	url, _ := url.ParseRequestURI(s.info.Address)
	url.Path = strings.Join([]string{url.Path, "/history/query"}, "")

	param := &QueryParam{Tags: tags, BeginTime: beginTime, EndTime: endTime, Count: valueCount}
	result := &QueryResult{}
	_, err = net.HTTPPost(s.httpClient, url.String(), param, result)
	if err != nil {
		return
	}

	if result.ErrorCode != 0 {
		err = fmt.Errorf("query history failed, erro:%s", result.Reason)
		return
	}

	ret = result.Data.Values

	return
}

func (s *piImpl) checkHealth() (err error) {
	url, _ := url.ParseRequestURI(s.info.Address)
	url.Path = strings.Join([]string{url.Path, "/ishealth"}, "")

	result := &CheckHealthResult{}
	_, err = net.HTTPGet(s.httpClient, url.String(), result)
	if err != nil {
		return
	}

	if result.ErrorCode != 0 {
		err = fmt.Errorf("check health failed, erro:%s", result.Reason)
		return
	}

	if result.Data.Status != 0 {
		err = fmt.Errorf("check health failed, invalid status:%d", result.Data.Status)
		return
	}

	return
}
