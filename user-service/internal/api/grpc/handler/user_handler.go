package handler

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"user-service/internal/domain/model"
	"user-service/internal/domain/services"
	pb "user-service/user-service/proto/user"
)

// UserHandler implements UserServiceServer interface
type UserHandler struct {
	pb.UnimplementedUserServiceServer
	userService services.UserService
}

// NewUserHandler creates new user handler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles user creation request
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Validate request
	if err := h.validateCreateUserRequest(req); err != nil {
		return &pb.CreateUserResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto request to service request
	serviceReq := &services.CreateUserRequest{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
	}

	// Call service layer
	user, err := h.userService.CreateUser(ctx, serviceReq)
	if err != nil {
		return &pb.CreateUserResponse{
			Success: false,
			Message: err.Error(),
		}, h.handleServiceError(err)
	}

	// Convert domain model to proto response
	return &pb.CreateUserResponse{
		Success: true,
		Message: "User created successfully",
		User:    h.domainUserToProto(user),
	}, nil
}

// GetUser handles get user by ID request
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Validate request
	if req.Id == "" {
		return &pb.GetUserResponse{
			Success: false,
			Message: "User ID is required",
		}, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Call service layer
	user, err := h.userService.GetUser(ctx, req.Id)
	if err != nil {
		return &pb.GetUserResponse{
			Success: false,
			Message: err.Error(),
		}, h.handleServiceError(err)
	}

	// Convert domain model to proto response
	return &pb.GetUserResponse{
		Success: true,
		Message: "User retrieved successfully",
		User:    h.domainUserToProto(user),
	}, nil
}

// UpdateUser handles user update request
func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// Validate request
	if err := h.validateUpdateUserRequest(req); err != nil {
		return &pb.UpdateUserResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto request to service request
	serviceReq := &services.UpdateUserRequest{
		ID:          req.Id,
		Email:       req.Email,
		Username:    req.Username,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Status:      h.protoStatusToDomain(req.Status),
	}

	// Call service layer
	user, err := h.userService.UpdateUser(ctx, serviceReq)
	if err != nil {
		return &pb.UpdateUserResponse{
			Success: false,
			Message: err.Error(),
		}, h.handleServiceError(err)
	}

	// Convert domain model to proto response
	return &pb.UpdateUserResponse{
		Success: true,
		Message: "User updated successfully",
		User:    h.domainUserToProto(user),
	}, nil
}

// DeleteUser handles user deletion request
func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// Validate request
	if req.Id == "" {
		return &pb.DeleteUserResponse{
			Success: false,
			Message: "User ID is required",
		}, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Call service layer
	err := h.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		return &pb.DeleteUserResponse{
			Success: false,
			Message: err.Error(),
		}, h.handleServiceError(err)
	}

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// ListUsers handles list users request
func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Set default pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	// Convert proto request to service request
	serviceReq := &services.ListUsersRequest{
		Page:   int(page),
		Limit:  int(limit),
		Search: req.Search,
		Status: h.protoStatusToDomain(req.Status),
	}

	// Call service layer
	users, total, err := h.userService.ListUsers(ctx, serviceReq)
	if err != nil {
		return &pb.ListUsersResponse{
			Success: false,
			Message: err.Error(),
		}, h.handleServiceError(err)
	}

	// Convert domain models to proto
	protoUsers := make([]*pb.User, len(users))
	for i, user := range users {
		protoUsers[i] = h.domainUserToProto(user)
	}

	return &pb.ListUsersResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Users:   protoUsers,
		Total:   int32(total),
		Page:    page,
		Limit:   limit,
	}, nil
}

// domainUserToProto converts domain User to proto User
func (h *UserHandler) domainUserToProto(user *model.User) *pb.User {
	return &pb.User{
		Id:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Status:      h.domainStatusToProto(user.Status),
		CreatedAt:   timestamppb.New(user.CreatedAt),
		UpdatedAt:   timestamppb.New(user.UpdatedAt),
	}
}

// domainStatusToProto converts domain UserStatus to proto UserStatus
func (h *UserHandler) domainStatusToProto(status model.UserStatus) pb.UserStatus {
	switch status {
	case model.UserStatusActive:
		return pb.UserStatus_ACTIVE
	case model.UserStatusInactive:
		return pb.UserStatus_INACTIVE
	case model.UserStatusSuspended:
		return pb.UserStatus_SUSPENDED
	default:
		return pb.UserStatus_ACTIVE
	}
}

// protoStatusToDomain converts proto UserStatus to domain UserStatus
func (h *UserHandler) protoStatusToDomain(status pb.UserStatus) model.UserStatus {
	switch status {
	case pb.UserStatus_ACTIVE:
		return model.UserStatusActive
	case pb.UserStatus_INACTIVE:
		return model.UserStatusInactive
	case pb.UserStatus_SUSPENDED:
		return model.UserStatusSuspended
	default:
		return model.UserStatusActive
	}
}

// validateCreateUserRequest validates create user request
func (h *UserHandler) validateCreateUserRequest(req *pb.CreateUserRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if req.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	return nil
}

// validateUpdateUserRequest validates update user request
func (h *UserHandler) validateUpdateUserRequest(req *pb.UpdateUserRequest) error {
	if req.Id == "" {
		return fmt.Errorf("user ID is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	return nil
}

// handleServiceError converts service errors to gRPC errors
func (h *UserHandler) handleServiceError(err error) error {
	switch {
	case err == model.ErrUserNotFound:
		return status.Error(codes.NotFound, err.Error())
	case err == model.ErrEmailExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case err == model.ErrUsernameExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case err == model.ErrEmailRequired:
		return status.Error(codes.InvalidArgument, err.Error())
	case err == model.ErrUsernameRequired:
		return status.Error(codes.InvalidArgument, err.Error())
	case err == model.ErrFullNameRequired:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "Internal server error")
	}
}
