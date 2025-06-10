package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

func (u *User) CreatePartnerRegistration(ctx context.Context, req *pb.PartnerRegistration) (*common.Empty, error) {
	if req.UserId == "" {
		return nil, errors.New(utils.E_invalid_user_id)
	}
	user, err := u.Db.GetUser(&pb.UserRequest{Id: req.UserId})
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(utils.E_not_found_user)
	}
	pr, err := u.Db.GetPartnerRegistration(&pb.PartnerRegistrationRequest{UserId: req.UserId})
	if err != nil {
		return nil, err
	}
	log.Println("CreatePartnerRegistration:", req, "User:", user.GetId(), "PartnerRegistration:", pr)
	if pr != nil {
		if pr.GetState() == pb.PartnerRegistration_approved.String() {
			return nil, errors.New(utils.E_already_partner)
		}
		if pr.GetState() == pb.PartnerRegistration_pending.String() {
			return nil, errors.New(utils.E_seller_application_pending)
		}
	}
	req.Id = utils.MakePartnerRegistrationId()
	req.CreatedAt = time.Now().Unix()
	req.State = pb.PartnerRegistration_pending.String()
	if err := u.Db.CreatePartnerRegistration(req); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) GetPartnerRegistration(ctx context.Context, req *pb.PartnerRegistrationRequest) (*pb.PartnerRegistration, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	registration, err := u.Db.GetPartnerRegistration(req)
	if err != nil {
		return nil, err
	}
	if registration == nil {
		return nil, errors.New(utils.E_not_found_partner_registration)
	}
	return registration, nil
}

func (u *User) ListPartnerRegistration(ctx context.Context, req *pb.PartnerRegistrationRequest) (*pb.PartnerRegistrations, error) {
	log.Println("ListPartnerRegistration:", req)
	list, err := u.Db.ListPartnerRegistration(req)
	if err != nil {
		log.Println("ListPartnerRegistration error:", err)
		return nil, err
	}
	return &pb.PartnerRegistrations{PartnerRegistrations: list, Total: int32(len(list))}, nil
}

func (u *User) UpdatePartnerRegistration(ctx context.Context, req *pb.PartnerRegistration) (*common.Empty, error) {
	log.Println("UpdatePartnerRegistration:", req)
	if req.GetId() == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	registration, err := u.Db.GetPartnerRegistration(&pb.PartnerRegistrationRequest{
		Id: req.GetId(),
	})
	if registration == nil {
		return nil, errors.New(utils.E_not_found_partner_registration)
	}
	if err != nil {
		return nil, err
	}
	if registration.GetState() != pb.PartnerRegistration_pending.String() {
		return nil, errors.New(utils.E_invalid_state)
	}
	req.UpdatedAt = time.Now().Unix()
	if err := u.Db.UpdatePartnerRegistration(req, &pb.PartnerRegistration{UserId: req.GetUserId()}); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) DeletePartnerRegistration(ctx context.Context, req *pb.PartnerRegistration) (*common.Empty, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	if err := u.Db.DeletePartnerRegistration(req); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) UpdateStatePartnerRegistration(ctx context.Context, req *pb.PartnerRegistration) (*common.Empty, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	registration, err := u.Db.GetPartnerRegistration(&pb.PartnerRegistrationRequest{
		Id: req.GetId(),
	})
	if registration == nil {
		return nil, errors.New(utils.E_not_found_partner_registration)
	}
	if err != nil {
		return nil, err
	}
	registration.UpdatedAt = time.Now().Unix()
	if req.GetState() == pb.PartnerRegistration_approved.String() {
		registration.State = pb.PartnerRegistration_approved.String()
		if err := u.Db.TranApprovePartnerRegistration(registration); err != nil {
			return nil, err
		}
	} else if req.GetState() == pb.PartnerRegistration_rejected.String() {
		registration.State = pb.PartnerRegistration_rejected.String()
		registration.ReasonReject = req.GetReasonReject()
		if err := u.Db.UpdatePartnerRegistration(registration, &pb.PartnerRegistration{Id: req.GetId()}); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(utils.E_invalid_state)
	}
	if err := u.SendEmailPartnerRegistrationStatus(registration); err != nil {
		log.Println("send email err:", err)
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) SendEmailPartnerRegistrationStatus(re *pb.PartnerRegistration) error {
	user, err := u.Db.GetUser(&pb.UserRequest{Id: re.GetUserId()})
	if err != nil {
		log.Println("get user err:", err)
		return err
	}
	var bin []byte
	var subject string
	log.Println("Gửi mail:", user.GetEmail())
	if re.GetState() == pb.PartnerRegistration_approved.String() {
		subject = "🎉 Chúc mừng! Đơn đăng ký bán hàng của bạn đã được duyệt"
		bin, err = os.ReadFile("assets/approved_partner.html")
		if err != nil {
			log.Println("read file err:", err)
			return err
		}
	} else if re.GetState() == pb.PartnerRegistration_rejected.String() {
		subject = "😔 Thông báo về đơn đăng ký bán hàng của bạn"
		bin, err = os.ReadFile("assets/rejected_partner.html")
		if err != nil {
			log.Println("read file err:", err)
			return err
		}
	} else {
		return errors.New("invalid state for sending email")
	}
	sending_time, _ := utils.ConvertUnixToDateTime("2006-01-02 15:04:05", time.Now().Unix())
	bodyMail := string(bin)
	metric := map[string]string{
		"fullname":      user.GetFullName(),
		"createdAt":     sending_time,
		"reason_reject": re.GetReasonReject(),
	}
	for k, v := range metric {
		bodyMail = strings.Replace(bodyMail, "{{"+k+"}}", v, -1)
	}
	err = utils.SendEmailPartnerRegistration(
		config.MailKey,
		config.MailUrl,
		user.GetEmail(),
		re.GetState(),
		subject,
		bodyMail,
	)
	return err
}
