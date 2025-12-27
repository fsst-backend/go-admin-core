package response

type Response struct {
	// 数据集
	RequestId string `protobuf:"bytes,1,opt,name=requestId,proto3" json:"requestId"`
	Code      int32  `protobuf:"varint,2,opt,name=code,proto3" json:"code"`
	Msg       string `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg"`
	Status    string `protobuf:"bytes,4,opt,name=status,proto3" json:"status"`
}

type response struct {
	Response
	Message interface{} `json:"message"`
}

type Page struct {
	Count  int `json:"count"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type page struct {
	Page
	List interface{} `json:"list"`
}

func (e *response) SetData(data interface{}) {
	e.Message = data
}

func (e response) Clone() Responses {
	return &e
}

func (e *response) SetTraceID(id string) {
	e.RequestId = id
}

func (e *response) SetMsg(s string) {
	e.Msg = s
}

func (e *response) SetCode(code int32) {
	e.Code = code
}

func (e *response) SetSuccess(success bool) {
	if !success {
		e.Status = "error"
	}
}
