package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"
	pb "pvz-service/proto"

	"google.golang.org/grpc"
)

type Server = grpc.Server

type PVZServer struct {
	pb.UnimplementedPVZServiceServer
	pvzService interfaces.PVZService
}

func NewPVZServer(pvzService interfaces.PVZService) *PVZServer {
	return &PVZServer{
		pvzService: pvzService,
	}
}

func (s *PVZServer) ListPVZ(ctx context.Context, req *pb.ListPVZRequest) (*pb.ListPVZResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("получен gRPC запрос на получение списка ПВЗ")

	options := models.PVZListOptions{
		Page:  1,
		Limit: 10000,
	}

	pvzs, total, err := s.pvzService.ListPVZ(ctx, options)
	if err != nil {
		log.Error("ошибка получения списка ПВЗ через gRPC", "error", err)
		return nil, err
	}

	response := &pb.ListPVZResponse{
		Items: make([]*pb.PVZ, 0, len(pvzs)),
	}

	for _, pvzWithReceptions := range pvzs {
		pvz := pvzWithReceptions.PVZ
		response.Items = append(response.Items, &pb.PVZ{
			Id:               pvz.ID.String(),
			RegistrationDate: pvz.RegistrationDate.Format(time.RFC3339),
			City:             pvz.City,
		})
	}

	log.Info("gRPC успешно отправлен список ПВЗ", "count", len(response.Items), "total", total)
	return response, nil
}

func StartGRPCServer(pvzService interfaces.PVZService, port int) *Server {
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("failed to listen on port %d: %v\n", port, err)
		return nil
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPVZServiceServer(grpcServer, NewPVZServer(pvzService))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Printf("failed to serve gRPC: %v\n", err)
		}
	}()

	return grpcServer
}
