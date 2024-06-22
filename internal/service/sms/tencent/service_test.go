package tencent

import (
	"context"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"testing"
)

func TestService_Send(t *testing.T) {
	type fields struct {
		client   *sms.Client
		appId    *string
		signName *string
	}
	type args struct {
		ctx     context.Context
		tplId   string
		args    []string
		numbers []string
	}
	var tests []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				client:   tt.fields.client,
				appId:    tt.fields.appId,
				signName: tt.fields.signName,
			}
			if err := s.Send(tt.args.ctx, tt.args.tplId, tt.args.args, tt.args.numbers...); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
