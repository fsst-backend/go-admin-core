package response

type ResponseMeta struct {
	// 数据集
	RequestId string `protobuf:"bytes,1,opt,name=requestId,proto3" json:"requestId"`
	Code      int32  `protobuf:"varint,2,opt,name=code,proto3" json:"code"`
	Msg       string `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg"`
	Status    string `protobuf:"bytes,4,opt,name=status,proto3" json:"status"`
}

type Response struct {
	ResponseMeta
	// 实体返回放到 message 字段，不再使用 data
	Message any `json:"message"`
}

type PageMeta struct {
	Count  int `json:"count"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type Page struct {
	PageMeta
	List any `json:"list"`
}

func (e *Response) SetData(data any) {
	e.Message = data
}

func (e *Response) SetTraceID(id string) {
	e.RequestId = id
}

func (e *Response) SetMsg(s string) {
	e.Msg = s
}

func (e *Response) SetCode(code int32) {
	e.Code = code
}

func (e *Response) SetSuccess(success bool) {
	if !success {
		e.Status = "error"
	}
}
