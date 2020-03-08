package proxy

import (
	"errors"
	"io"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	model "supos.ai/data-lake/external/tsdb-proxy/common/model"
	pb "supos.ai/data-lake/external/tsdb-proxy/protocol/supos_edges"
)

// Callback callback
type Callback interface {
	OnNotifyCallBack(msg *pb.DownChannel)
}

// RtdProxy rtdservice proxy
type RtdProxy interface {
	Start(callBack Callback) error
	Stop()

	UpdateContextMeta(key, val string)

	Invoke(upMsg *pb.UpChannel) error
	CreateObject(obj *model.MetaObject) (int32, error)
	AddTag(objectName string, tags []*model.MetaProperty) (int32, error)
	DelTag(objectName string, tags []string) (int32, error)
	ModTag(objectName string, tags []*model.MetaProperty) (int32, error)
	UpdateTagValue(values *model.ValueSequnce) error
}

type rtdProxy struct {
	rtdAddress string

	callBack                        Callback
	clientConnect                   *grpc.ClientConn
	dataCollectorServiceClient      pb.DataCollectorServiceClient
	collectorToServiceNotifyChannel pb.DataCollectorService_CollectorSerivceChannelClient
	collectorToServiceDataChannel   pb.DataCollectorService_UpdateTagValueClient

	runningFlag bool

	sessionID                   string
	dataCollectorServiceContext context.Context
}

// NewProxy create new proxy
func NewProxy(rtdAddress string) RtdProxy {

	return &rtdProxy{rtdAddress: rtdAddress}
}

func (s *rtdProxy) Start(callBack Callback) error {
	conn, err := grpc.Dial(s.rtdAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("dial to rtdService failed, err:%s", err.Error())
		return err
	}

	dataCollectorServiceClient := pb.NewDataCollectorServiceClient(conn)

	dataCollectorServiceContext := context.Background()
	s.dataCollectorServiceContext = dataCollectorServiceContext

	collectorToServiceNotifyChannel, err := dataCollectorServiceClient.CollectorSerivceChannel(s.dataCollectorServiceContext)
	if err != nil {
		log.Printf("create Collector to Service Notify Channel failed, rtdAddress:%s, err:%s", s.rtdAddress, err.Error())
		return err
	}

	s.clientConnect = conn
	s.dataCollectorServiceClient = dataCollectorServiceClient
	s.collectorToServiceNotifyChannel = collectorToServiceNotifyChannel
	s.callBack = callBack

	go s.run()

	return nil
}

func (s *rtdProxy) Stop() {
	s.runningFlag = false

	if s.clientConnect != nil {
		if s.collectorToServiceNotifyChannel != nil {
			s.collectorToServiceNotifyChannel.CloseSend()
			s.collectorToServiceNotifyChannel = nil
		}

		if s.collectorToServiceDataChannel != nil {
			s.collectorToServiceDataChannel.CloseSend()
			s.collectorToServiceDataChannel = nil
		}

		s.dataCollectorServiceClient = nil

		s.clientConnect.Close()
		s.clientConnect = nil
	}

	s.callBack = nil
}

func (s *rtdProxy) UpdateContextMeta(key, val string) {
	if s.dataCollectorServiceContext != nil {
		s.dataCollectorServiceContext = metadata.AppendToOutgoingContext(s.dataCollectorServiceContext, key, val)
	}
}

func (s *rtdProxy) run() {
	s.runningFlag = true

	log.Printf("rtdProxy running...")
	for s.runningFlag {
		downMsg, err := s.collectorToServiceNotifyChannel.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("collectorToServiceNotifyChannel.Recv, err:%s", err.Error())
			panic(err)
		}

		if s.callBack != nil {
			s.callBack.OnNotifyCallBack(downMsg)
		}
	}

	log.Printf("rtdProxy terminate...")
}

func (s *rtdProxy) Invoke(upMsg *pb.UpChannel) error {
	if s.collectorToServiceNotifyChannel == nil {
		return errors.New("illegal collectorToServiceNotifyChannel")
	}

	return s.collectorToServiceNotifyChannel.Send(upMsg)
}

func (s *rtdProxy) CreateObject(obj *model.MetaObject) (int32, error) {

	request := &pb.MetaDataRequest{}
	opt := &pb.MetaDataOperation{Action: pb.MetaDataAction_MetaDataAction_ADD}
	pro := &model.PropertyOrObject{Filed: &model.PropertyOrObject_Obj{Obj: obj}}
	opt.AddOrModify = pro

	request.Operations = append(request.Operations, opt)

	rsp, err := s.dataCollectorServiceClient.DoMetaData(s.dataCollectorServiceContext, request)
	if err != nil {
		return -1, err
	}
	return rsp.GetErrorCode(), nil
}

func (s *rtdProxy) AddTag(objectName string, tags []*model.MetaProperty) (int32, error) {
	request := &pb.MetaDataRequest{}
	for _, val := range tags {
		opt := &pb.MetaDataOperation{Action: pb.MetaDataAction_MetaDataAction_ADD, ObjectName: objectName}
		pro := &model.PropertyOrObject{Filed: &model.PropertyOrObject_Prop{Prop: val}}
		opt.AddOrModify = pro

		request.Operations = append(request.Operations, opt)
	}

	rsp, err := s.dataCollectorServiceClient.DoMetaData(s.dataCollectorServiceContext, request)
	if err != nil {
		return -1, err
	}
	return rsp.GetErrorCode(), nil
}

func (s *rtdProxy) ModTag(objectName string, tags []*model.MetaProperty) (int32, error) {
	request := &pb.MetaDataRequest{}
	for _, val := range tags {
		opt := &pb.MetaDataOperation{Action: pb.MetaDataAction_MetaDataAction_MODIFY, ObjectName: objectName}
		pro := &model.PropertyOrObject{Filed: &model.PropertyOrObject_Prop{Prop: val}}
		opt.AddOrModify = pro

		request.Operations = append(request.Operations, opt)
	}

	rsp, err := s.dataCollectorServiceClient.DoMetaData(s.dataCollectorServiceContext, request)
	if err != nil {
		return -1, err
	}
	return rsp.GetErrorCode(), nil
}

func (s *rtdProxy) DelTag(objectName string, tags []string) (int32, error) {
	request := &pb.DeleteTagRequest{Tags: tags}

	rsp, err := s.dataCollectorServiceClient.DeleteTags(s.dataCollectorServiceContext, request)
	if err != nil {
		return -1, err
	}
	return rsp.GetErrorCode(), nil
}

func (s *rtdProxy) UpdateTagValue(values *model.ValueSequnce) error {
	if s.collectorToServiceDataChannel == nil {
		collectorToServiceDataChannel, err := s.dataCollectorServiceClient.UpdateTagValue(s.dataCollectorServiceContext)
		if err != nil {
			log.Printf("create Collector to Service Data Channel failed, err:%s", err.Error())
			return err
		}

		s.collectorToServiceDataChannel = collectorToServiceDataChannel
	}

	err := s.collectorToServiceDataChannel.Send(values)
	return err
}
