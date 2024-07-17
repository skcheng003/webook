package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/skcheng003/webook/internal/repository"
	"github.com/skcheng003/webook/internal/service/sms"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

// const codeTplId = "1877556"
const codeTplId = "1110"

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type SMSCodeService struct {
	sms  sms.Service
	repo repository.CodeRepository
}

func NewSMSCodeService(svc sms.Service, repo repository.CodeRepository) CodeService {
	return &SMSCodeService{
		sms:  svc,
		repo: repo,
	}
}

// Send 发送验证码，用 biz 来对业务场景进行区分
func (svc *SMSCodeService) Send(ctx context.Context, biz string, phone string) error {
	code := svc.generate()
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	err = svc.sms.Send(ctx, codeTplId, []string{code}, phone)
	// 如果 err != nil, 可以考虑设计一个 retrySendService 来进行重试，不管也行
	return err
}

// Verify 对验证码进行验证
func (svc *SMSCodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	// 这里我们在 service 层面上对 Handler 屏蔽了最为特殊的错误
	if errors.Is(err, ErrCodeSendTooMany) {
		// 在接入了告警之后，这边要告警
		// 因为这意味着有人在搞你
		return false, nil
	}
	return ok, err
}

func (svc *SMSCodeService) generate() string {
	num := rand.Intn(999999)
	return fmt.Sprintf("%06d", num)
}
