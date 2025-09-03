package main

import (
	"fmt"
	"log"
	"net"

	"github.com/endyapina/muzzapp/internal/config"
	"github.com/endyapina/muzzapp/internal/database"
	"github.com/endyapina/muzzapp/internal/redis"
	"github.com/endyapina/muzzapp/internal/repository"
	"github.com/endyapina/muzzapp/internal/service"
	"github.com/endyapina/muzzapp/internal/web"

	pb "github.com/endyapina/muzzapp/proto/gen/github.com/muzzapp/backend-interview-task"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to connect to DB: %w", err))
	}
	log.Println("database connection successful...")

	cache, err := redis.NewCache(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create redis cache: %w", err))
	}
	log.Println("redis cache connection successful...")

	repo, err := repository.New(db, cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create database repository: %w", err))
	}
	service := service.New(repo, cache)
	handler := web.NewHandler(service)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterExploreServiceServer(grpcServer, handler)

	log.Printf("gRPC Server running on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
