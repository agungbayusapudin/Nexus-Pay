# gRPC Handler Documentation

## Overview
File `user_handler.go` adalah implementasi utama dari gRPC service yang mengimplementasikan interface `UserServiceServer` yang di-generate dari proto file. Handler ini bertanggung jawab untuk:
- Menerima gRPC requests
- Validasi input
- Konversi data antara proto dan domain models
- Memanggil business logic layer
- Mengembalikan response yang sesuai

## Code Structure

### Package Declaration & Imports
```go
package handler

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	
	pb "user-service/proto/user"
	"user-service/internal/domain/model"
	"user-service/internal/domain/services"
)
```

**Penjelasan Imports:**
- `context`: Untuk handling request context dan cancellation
- `fmt`: Untuk string formatting dalam error messages
- `time`: Untuk timestamp operations
- `google.golang.org/grpc/codes`: gRPC status codes (NotFound, InvalidArgument, dll)
- `google.golang.org/grpc/status`: Untuk membuat gRPC error responses
- `timestamppb`: Untuk konversi Go time.Time ke protobuf Timestamp
- `pb`: Alias untuk generated protobuf types
- `model`: Domain models (User, UserStatus, dll)
- `services`: Business logic interfaces

### Handler Struct
```go
type UserHandler struct {
	pb.UnimplementedUserServiceServer
	userService services.UserService
}
```

**Penjelasan:**
- `pb.UnimplementedUserServiceServer`: Embedded struct dari generated code yang menyediakan default implementation untuk semua methods. Ini memastikan forward compatibility jika proto file ditambah method baru.
- `userService`: Dependency injection untuk business logic layer

### Constructor
```go
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
```

**Penjelasan:**
- Constructor pattern untuk dependency injection
- Menerima service layer sebagai parameter
- Return pointer ke UserHandler instance

## gRPC Methods Implementation

### CreateUser Method
```go
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
```

**Method Signature:**
- `ctx context.Context`: Request context untuk timeout, cancellation, metadata
- `req *pb.CreateUserRequest`: Proto request message dari client
- Return: `(*pb.CreateUserResponse, error)` - Proto response dan error

#### Step 1: Request Validation
```go
if err := h.validateCreateUserRequest(req); err != nil {
	return &pb.CreateUserResponse{
		Success: false,
		Message: err.Error(),
	}, status.Error(codes.InvalidArgument, err.Error())
}
```

**Penjelasan:**
- Validasi input sebelum processing
- Return response dengan `Success: false` untuk client
- Return gRPC error dengan `codes.InvalidArgument` untuk proper error handling
- Dual return: response untuk business logic, error untuk gRPC layer

#### Step 2: Proto to Service Conversion
```go
serviceReq := &services.CreateUserRequest{
	Email:       req.Email,
	Username:    req.Username,
	Password:    req.Password,
	FullName:    req.FullName,
	PhoneNumber: req.PhoneNumber,
}
```

**Penjelasan:**
- Konversi dari proto message ke service layer DTO
- Memisahkan concerns: proto layer vs business layer
- Service layer tidak tahu tentang gRPC/proto details

#### Step 3: Business Logic Call
```go
user, err := h.userService.CreateUser(ctx, serviceReq)
if err != nil {
	return &pb.CreateUserResponse{
		Success: false,
		Message: err.Error(),
	}, h.handleServiceError(err)
}
```

**Penjelasan:**
- Delegate ke service layer untuk business logic
- Pass context untuk timeout/cancellation propagation
- Error handling dengan custom error mapper
- `handleServiceError()` converts domain errors ke gRPC status codes

#### Step 4: Success Response
```go
return &pb.CreateUserResponse{
	Success: true,
	Message: "User created successfully",
	User:    h.domainUserToProto(user),
}, nil
```

**Penjelasan:**
- Success response dengan user data
- Convert domain model ke proto message
- Return `nil` error untuk success case

### GetUser Method
```go
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
```

#### Input Validation
```go
if req.Id == "" {
	return &pb.GetUserResponse{
		Success: false,
		Message: "User ID is required",
	}, status.Error(codes.InvalidArgument, "User ID is required")
}
```

**Penjelasan:**
- Simple validation untuk required field
- Immediate return jika validation gagal
- Consistent error response pattern

### UpdateUser Method
Similar pattern dengan CreateUser, tapi dengan update-specific validation dan logic.

### DeleteUser Method
```go
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
```

**Penjelasan:**
- Delete operation tidak return data, hanya success/failure
- Simple success response tanpa user data

### ListUsers Method
```go
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
```

**Penjelasan:**
- Default pagination values untuk user experience
- Max limit untuk prevent abuse
- Business rules di handler level

## Helper Methods

### Domain to Proto Conversion
```go
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
```

**Penjelasan:**
- Mapping dari domain model ke proto message
- `timestamppb.New()` untuk convert Go time.Time ke protobuf Timestamp
- Status conversion dengan separate method untuk reusability

### Status Conversion
```go
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
```

**Penjelasan:**
- Explicit mapping antara domain enums dan proto enums
- Default case untuk safety
- Bidirectional conversion (proto to domain dan sebaliknya)

### Validation Methods
```go
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
```

**Penjelasan:**
- Input validation dengan clear error messages
- Business rules (password length)
- Early validation sebelum hit service layer

### Error Handling
```go
func (h *UserHandler) handleServiceError(err error) error {
	switch {
	case err == model.ErrUserNotFound:
		return status.Error(codes.NotFound, err.Error())
	case err == model.ErrEmailExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case err == model.ErrUsernameExists:
		return status.Error(codes.AlreadyExists, err.Error())
	// ... more cases
	default:
		return status.Error(codes.Internal, "Internal server error")
	}
}
```

**Penjelasan:**
- Mapping domain errors ke gRPC status codes
- Proper HTTP-like status codes untuk clients
- Default case untuk unknown errors
- Security: Hide internal error details dari clients

## Best Practices Implemented

1. **Separation of Concerns**: Handler hanya handle gRPC concerns, business logic di service layer
2. **Error Handling**: Proper gRPC error codes dengan meaningful messages
3. **Validation**: Input validation di handler level
4. **Type Safety**: Strong typing dengan proto dan domain models
5. **Context Propagation**: Pass context ke service layer untuk timeout/cancellation
6. **Defensive Programming**: Default values, max limits, nil checks
7. **Clean Code**: Clear method names, consistent patterns, good documentation
