package main

import (
	"log"
	"testing"

	pb "github.com/huyshop/header/user"
)

func Test_send(t *testing.T) {

	u, err := NewUser(config)
	if err != nil {
		log.Println("Error creating User instance:", err)
	}
	err = u.SendEmailPartnerRegistrationStatus(&pb.PartnerRegistration{
		Id:     "test-id",
		UserId: "userd0oosau9ipffg4e7ag1g",
		State:  pb.PartnerRegistration_rejected.String(),
		ReasonReject: "Test reason for rejection",
	})
	if err != nil {
		t.Errorf("SendEmailPartnerRegistrationStatus failed: %v", err)
	} else {
		t.Logf("SendEmailPartnerRegistrationStatus succeeded")
	}
}
