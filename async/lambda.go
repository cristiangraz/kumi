package async

import (
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/cristiangraz/kumi/api"
)

// LambdaInvoker executes lambda functions.
type LambdaInvoker struct {
	svc *lambda.Lambda
}

// NewLambdaInvoker returns a new LambdaInvoker.
func NewLambdaInvoker(svc *lambda.Lambda) *LambdaInvoker {
	return &LambdaInvoker{svc}
}

// Invoke executes a lambda function. If async is set to false, Invoke will
// return the api response from lambda.
// @todo add configurable option for async methods to be invoked via SNS.
func (l *LambdaInvoker) Invoke(name string, msg *Message, async bool) (*api.Response, error) {
	invocationType := "Event"
	if async == false {
		invocationType = "RequestResponse"
	}
	params := &lambda.InvokeInput{
		FunctionName:   aws.String(name),
		InvocationType: aws.String(invocationType),
		Payload:        msg.Payload,
	}
	if len(msg.Context) > 0 {
		params.ClientContext = aws.String(base64.StdEncoding.EncodeToString(msg.Context))
	}

	resp, err := l.svc.Invoke(params)
	if err != nil {
		return nil, err
	}

	if async {
		return nil, nil
	}

	var r api.Response
	err = json.Unmarshal(resp.Payload, &r)

	return &r, err
}
