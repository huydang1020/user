package main

import (
	"context"
	"strconv"

	pb "github.com/huyshop/header/user"
)

func (u *User) GetReportUser(ctx context.Context, rq *pb.ReportRequest) (*pb.ReportUsers, error) {
	from, _ := strconv.ParseInt(rq.From, 10, 64)
	to, _ := strconv.ParseInt(rq.To, 10, 64)
	users, err := u.Db.ListUsers(&pb.UserRequest{
		From: from,
		To:   to,
	})
	if err != nil {
		return nil, err
	}

	reportUsers := make([]*pb.ReportUser, 0, len(users))
	for _, user := range users {
		// Tính tổng điểm đã tích lũy (cộng) và đã tiêu (trừ) từ point_exchange
		pointExchanges, _ := u.Db.ListPointExchange(&pb.PointExchangeRequest{ReceiverId: user.Id})
		totalPointsEarned := int64(0)
		totalPointsSpent := int64(0)
		for _, ex := range pointExchanges {
			if ex.Points > 0 {
				totalPointsEarned += ex.Points
			} else if ex.Points < 0 {
				totalPointsSpent += -ex.Points
			}
		}

		reportUsers = append(reportUsers, &pb.ReportUser{
			UserId:            user.Id,
			FullName:          user.FullName,
			PhoneNumber:       user.PhoneNumber,
			TotalPointsEarned: totalPointsEarned,
			TotalPointsSpent:  totalPointsSpent,
		})
	}

	return &pb.ReportUsers{
		Users: reportUsers,
		Total: int32(len(reportUsers)),
	}, nil
}
