package jsend

// jsend statuses
type Status string

const (
    SuccessStatus   Status = "success"
    FailStatus      Status = "fail"
    ErrorStatus     Status = "error"
)

// jsend response
type Response struct {
    Status      Status  `json:"status"`
    Data        any     `json:"data,omitempty"`
    Message     *string `json:"message,omitempty"`
    Code        *int    `json:"code,omitempty"` 
    HttpCode    int     `json:"-"`
}

// create jsend success response
func Success(data any, err *string, code *int, httpCode int) *Response {
    return &Response{
        Status:     SuccessStatus,
        Data:       data,
        Message:    err,
        Code:       code,
        HttpCode:   httpCode,
    }
}

// create jsend fail response
func Fail(data any, err string, code *int, httpCode int) *Response {
    return &Response{
        Status:     FailStatus,
        Data:       data,
        Message:    &err,
        Code:       code,
        HttpCode:   httpCode,
    }
}

// create jsend error response 
func Error(data any, err string, code *int, httpCode int) *Response {
    return &Response{
        Status:     ErrorStatus,
        Data:       data,
        Message:    &err,
        Code:       code,
        HttpCode:   httpCode,
    }
}
