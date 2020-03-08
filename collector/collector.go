package collector

import (
	"errors"
	"fmt"
	"log"

	"supos.ai/data-lake/external/tsdb-proxy/common/model"
	"supos.ai/data-lake/external/tsdb-proxy/proxy"
	"supos.ai/data-lake/external/tsdb-proxy/snapshot"

	pb "supos.ai/data-lake/external/tsdb-proxy/protocol/supos_edges"
)

const (
	// Init initialize
	Init = iota
	// LoginFailed login failed
	LoginFailed
	//Ready login already
	Ready
	// Alive submit meta alreay
	Alive
	//TimeOut timeout
	TimeOut
	//Destroy destory
	Destroy
	//KickOff kickoff
	KickOff
)

const (
	// LoginStatus login status
	LoginStatus = iota
	// TerminateStatus terminate status
	TerminateStatus
)

// StatusCallBack statusCallBack
type StatusCallBack func(collectName string, status, errorCode int, reason string)

// Collector 采集器
type Collector interface {
	Name() string

	Start(rtdService string, statusCallBack StatusCallBack) error
	Stop()

	Destroy()

	IsReady() bool

	UpdateValue(values *model.ValueSequnce) error
	UpdateMeta(meta []*model.MetaProperty) error
}

// NewCollector 新建Collector
func NewCollector(endPoint string) Collector {
	snapshot := snapshot.New(endPoint)
	fmt.Printf("new collect endpoint:%s", endPoint)

	return &collector{
		endPoint: endPoint,
		status:   Init,
		snapshot: snapshot,
		metas:    make(map[string]*model.MetaProperty),
	}
}

type collector struct {
	endPoint   string
	remoteAddr string
	loginTime  int64

	status         int
	metas          map[string]*model.MetaProperty
	snapshot       snapshot.Snapshot
	rtdProxy       proxy.RtdProxy
	statusCallBack StatusCallBack
}

func (s *collector) Name() string {
	return s.endPoint
}

func (s *collector) Start(rtdService string, statusCallBack StatusCallBack) error {
	proxy := proxy.NewProxy(rtdService)

	err := proxy.Start(s)
	if err != nil {
		log.Printf("start rtdProxy failed, rtddService:%s, err:%s", rtdService, err.Error())
		return err
	}

	s.rtdProxy = proxy
	s.statusCallBack = statusCallBack

	return s.login()
}

func (s *collector) Stop() {
	s.logout()

	if s.rtdProxy != nil {
		s.rtdProxy.Stop()
	}
}

func (s *collector) IsReady() bool {
	return s.status == Ready || s.status == Alive
}

func (s *collector) Destroy() {
	s.status = Destroy

	values := s.snapshot.GetAllCurrentTagValue()
	if values != nil {
		needSend := false
		notifyVal := &model.ValueSequnce{Value: []*model.NamedValue{}}
		for _, val := range values.GetValue() {
			_, ok := val.GetValue().GetKind().(*model.Value_PrimitiveValueWithQT)
			if ok {
				vqtVal := val.Value.GetPrimitiveValueWithQT()
				vqtVal.Status = 0x8000000000000000

				tagName := val.GetName()
				if tagName == "online" {
					vqtVal.Value = &model.PrimitiveValue{Value: &model.PrimitiveValue_BoolValue{BoolValue: false}}
				}

				needSend = true
				notifyVal.Value = append(notifyVal.Value, &model.NamedValue{Name: tagName, Value: &model.Value{Kind: &model.Value_PrimitiveValueWithQT{PrimitiveValueWithQT: vqtVal}}})
			}
		}

		if needSend {
			s.pushRtdData(notifyVal)
		}
	}

	s.Stop()

	s.snapshot = nil
}

func (s *collector) UpdateMeta(meta []*model.MetaProperty) error {
	if !s.IsReady() {
		return fmt.Errorf("invalid status")
	}

	newMap := make(map[string]*model.MetaProperty)

	add := make([]*model.MetaProperty, 0, 0)
	mod := make([]*model.MetaProperty, 0, 0)
	del := make([]string, 0, 0)

	for _, m := range meta {
		if _, ok := s.metas[m.Name]; ok {
			mod = append(mod, m)
		} else {
			add = append(add, m)

		}
		newMap[m.Name] = m
	}
	for _, m := range s.metas {
		if _, ok := newMap[m.Name]; !ok {
			del = append(del, m.Name)
		}
	}

	var err error
	for {
		if len(add) > 0 {
			err = s.addMetaTags(add)
			if err != nil {
				break
			}
		}
		if len(mod) > 0 {
			err = s.modMetaTag(mod)
			if err != nil {
				break
			}
		}
		if len(del) > 0 {
			err = s.delMetaTag(del)
			if err != nil {
				break
			}
		}

		break
	}

	return err
}

func (s *collector) UpdateValue(values *model.ValueSequnce) error {
	if !s.IsReady() {
		return fmt.Errorf("invalid status")
	}

	newValues := s.snapshot.UpdateTagValues(values)
	if s.status == Ready {
		obj := NewMetaObject(s.endPoint, newValues)

		err := s.createMetaObject(obj)
		if err != nil {
			return err
		}

		s.status = Alive
	}

	return s.pushRtdData(values)
}

func (s *collector) GetAllSnapshotValue() *model.ValueSequnce {
	return s.snapshot.GetAllCurrentTagValue()
}

func (s *collector) OnNotifyCallBack(msg *pb.DownChannel) {
	cmd := msg.GetCmd()
	loginResponse, ok := cmd.(*pb.DownChannel_LoginResponse)
	if ok {
		errCode := loginResponse.LoginResponse.GetErrorCode()
		reason := loginResponse.LoginResponse.GetReason()
		sessionID := loginResponse.LoginResponse.GetSessionID()

		if errCode == 100004 {
			s.rtdProxy.UpdateContextMeta("session", sessionID)

			s.status = Ready
		} else {
			s.status = LoginFailed
		}

		if s.statusCallBack != nil {
			s.statusCallBack(s.endPoint, LoginStatus, 0, reason)
		}

		log.Printf("errCode:%d, reason:%s, sessionID:%s, status:%d", errCode, reason, sessionID, s.status)
		return
	}

	_, ok = cmd.(*pb.DownChannel_LogoutResponse)
	if ok {
		return
	}
	notify, ok := cmd.(*pb.DownChannel_StatusNotify)
	if ok {
		if notify.StatusNotify.GetCmd() == pb.ResultCode_ServerKickOff {
			s.status = KickOff

			s.rtdProxy.UpdateContextMeta("session", "")
		}

		if s.statusCallBack != nil {
			s.statusCallBack(s.endPoint, TerminateStatus, 0, "supos terminate")
		}

		return
	}
}

func (s *collector) login() error {
	//special todo !!!!!!
	uuid := "620d7e6f-4083-4036-8e89-4dac8748ee01"
	identify := "stddata-service"

	login := &pb.LoginRequest{AuthToken: uuid, IdentifyID: identify, EndpointName: s.endPoint}
	upMsg := &pb.UpChannel{Cmd: &pb.UpChannel_LoginRequest{LoginRequest: login}}

	err := s.rtdProxy.Invoke(upMsg)
	if err != nil {
		log.Printf("invoke failed, err:%s", err.Error())
	}

	return err
}

func (s *collector) logout() error {
	logout := &pb.LogoutRequest{}
	upMsg := &pb.UpChannel{Cmd: &pb.UpChannel_LogoutRequest{LogoutRequest: logout}}

	err := s.rtdProxy.Invoke(upMsg)
	if err != nil {
		log.Printf("invoke failed, err:%s", err.Error())
	}

	return err
}

func (s *collector) createMetaObject(obj *model.MetaObject) error {
	errCode, err := s.rtdProxy.CreateObject(obj)
	if err != nil {
		log.Printf("CreateObject failed, err:%s", err.Error())
		return err
	}

	if errCode != pb.ResultCode_Success {
		errMsg := fmt.Sprintf("errCode:%d", errCode)
		err = errors.New(errMsg)
		log.Printf("CreateObject failed, err:%s", err.Error())
	}

	return nil
}

func (s *collector) delMetaTag(tags []string) error {
	errCode, err := s.rtdProxy.DelTag(s.endPoint, tags)
	if err != nil {
		log.Printf("delete tag failed, err:%s", err.Error())
		return err
	}

	if errCode != pb.ResultCode_Success {
		errMsg := fmt.Sprintf("errCode:%d", errCode)
		err = errors.New(errMsg)
		log.Printf("delete tag failed, err:%s", err.Error())
	}

	log.Printf("delete tag s,del len:%v", len(tags))

	return nil
}

func (s *collector) modMetaTag(tags []*model.MetaProperty) error {
	errCode, err := s.rtdProxy.ModTag(s.endPoint, tags)
	if err != nil {
		log.Printf("modify tag failed, err:%s", err.Error())
		return err
	}

	if errCode != pb.ResultCode_Success {
		errMsg := fmt.Sprintf("errCode:%d", errCode)
		err = errors.New(errMsg)
		log.Printf("modify tag failed, err:%s", err.Error())
	}

	log.Printf("mod tag s,mod len:%v", len(tags))

	return nil
}

func (s *collector) addMetaTags(tags []*model.MetaProperty) error {
	errCode, err := s.rtdProxy.AddTag(s.endPoint, tags)
	if err != nil {
		log.Printf("AddTag failed, err:%s", err.Error())
		return err
	}

	if errCode != pb.ResultCode_Success {
		errMsg := fmt.Sprintf("errCode:%d", errCode)
		err = errors.New(errMsg)
		log.Printf("AddTag failed, err:%s", err.Error())
	}

	log.Printf("add tag s,add len:%v", len(tags))

	return nil
}

func (s *collector) pushRtdData(values *model.ValueSequnce) error {
	err := s.rtdProxy.UpdateTagValue(values)
	if err != nil {
		log.Printf("updateTagValue failed, err:%s", err.Error())
		return err
	}
	log.Printf("update tag value s,len:%v", len(values.Value))

	return nil
}
