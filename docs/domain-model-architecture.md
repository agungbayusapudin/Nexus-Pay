# Domain Model Architecture Documentation

## Overview: Mengapa Model Begitu Kompleks?

Dalam Clean Architecture, **Domain Model** adalah jantung dari aplikasi yang berisi **business logic** dan **business rules**. Model yang kompleks ini diperlukan karena beberapa alasan penting:

## 1. **Separation of Concerns (Pemisahan Tanggung Jawab)**

### Problem yang Diselesaikan:
Tanpa pemisahan yang jelas, kita akan memiliki masalah seperti:
```go
// ❌ BAD: Mixing concerns
type User struct {
    ID       string `json:"id" protobuf:"bytes,1,opt,name=id" db:"id"`
    Email    string `json:"email" protobuf:"bytes,2,opt,name=email" db:"email"`
    Password string `json:"password" protobuf:"bytes,3,opt,name=password" db:"password_hash"`
}
```

**Masalah:**
- Proto tags, JSON tags, DB tags tercampur
- Password exposed di JSON response
- Tidak ada business logic
- Sulit untuk testing
- Tidak ada validation

### Solution: Multiple Models untuk Different Concerns

```go
// ✅ GOOD: Separated models
type User struct {              // Domain Model - Business Logic
    ID           string
    Email        string
    PasswordHash string         // Internal representation
    Status       UserStatus
}

type CreateUserRequest struct { // Service Layer DTO
    Email    string
    Password string             // Plain password untuk input
    FullName string
}

type UserResponse struct {      // API Response DTO
    ID       string             // No password exposed
    Email    string
    FullName string
}
```

## 2. **Layer Responsibilities Explained**

### **Domain Model (User struct)**
```go
type User struct {
    ID           string     `db:"id"`
    Email        string     `db:"email"`
    PasswordHash string     `json:"-" db:"password_hash"` // Hidden from JSON
    Status       UserStatus `db:"status"`
    CreatedAt    time.Time  `db:"created_at"`
    UpdatedAt    time.Time  `db:"updated_at"`
}
```

**Mengapa Perlu:**
- **Business Entity**: Representasi core business object
- **Database Mapping**: Tags untuk ORM/SQL mapping
- **Security**: Password hash, bukan plain password
- **Business Rules**: Methods seperti `IsActive()`, `CanLogin()`
- **Data Integrity**: Validation methods

### **Request DTOs (Data Transfer Objects)**
```go
type CreateUserRequest struct {
    Email       string `json:"email" validate:"required,email"`
    Username    string `json:"username" validate:"required,min=3,max=50"`
    Password    string `json:"password" validate:"required,min=6"`
    FullName    string `json:"full_name" validate:"required"`
    PhoneNumber string `json:"phone_number" validate:"omitempty,phone"`
}
```

**Mengapa Perlu:**
- **Input Validation**: Validation tags untuk automatic validation
- **API Contract**: Clear contract untuk client requests
- **Security**: Separate dari domain model (tidak ada ID, timestamps)
- **Flexibility**: Bisa berbeda dari domain model structure

### **Response DTOs**
```go
type UserResponse struct {
    ID          string     `json:"id"`
    Email       string     `json:"email"`
    Username    string     `json:"username"`
    FullName    string     `json:"full_name"`
    Status      UserStatus `json:"status"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    // No PasswordHash - Security!
}
```

**Mengapa Perlu:**
- **Security**: Tidak expose sensitive data (password hash)
- **API Consistency**: Consistent response format
- **Versioning**: Bisa change response tanpa affect domain model
- **Client Needs**: Only data yang dibutuhkan client

## 3. **Business Logic Methods**

### **Why in Domain Model?**
```go
// Business rules dalam domain model
func (u *User) IsActive() bool {
    return u.Status == UserStatusActive
}

func (u *User) CanLogin() bool {
    return u.Status == UserStatusActive
}

func (u *User) IsSuspended() bool {
    return u.Status == UserStatusSuspended
}
```

**Alasan:**
- **Domain-Driven Design**: Business logic belongs to domain
- **Encapsulation**: Data dan behavior dalam satu tempat
- **Reusability**: Logic bisa dipakai di berbagai layer
- **Testing**: Easy untuk unit test business rules

## 4. **Validation Strategy**

### **Multi-Level Validation**
```go
// 1. DTO Level Validation (Input)
func (r *CreateUserRequest) Validate() error {
    if r.Email == "" {
        return ErrEmailRequired
    }
    // ... more validation
}

// 2. Domain Level Validation (Business Rules)
func (u *User) Validate() error {
    if err := u.validateEmail(); err != nil {
        return err
    }
    // ... more business validation
}
```

**Mengapa Dua Level:**
- **DTO Validation**: Input sanitization, format checking
- **Domain Validation**: Business rules, complex validation
- **Defense in Depth**: Multiple layers of protection
- **Clear Responsibility**: Each layer validates what it cares about

## 5. **Error Handling Strategy**

### **Domain Errors dengan Codes**
```go
type DomainError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

var (
    ErrUserNotFound   = NewDomainError("USER_NOT_FOUND", "User not found")
    ErrEmailExists    = NewDomainError("EMAIL_EXISTS", "Email already exists")
    ErrUserSuspended  = NewDomainError("USER_SUSPENDED", "User account is suspended")
)
```

**Mengapa Error Codes:**
- **Client Handling**: Clients bisa handle specific errors
- **Internationalization**: Error codes bisa di-translate
- **API Consistency**: Consistent error format
- **Debugging**: Easy untuk track specific error types

## 6. **Relationship dengan gRPC Proto**

### **Apakah Harus Sama dengan Proto?**

**TIDAK! Dan ini alasannya:**

#### **Proto Definition (API Contract)**
```protobuf
message User {
  string id = 1;
  string email = 2;
  string username = 3;
  string full_name = 4;
  string phone_number = 5;
  UserStatus status = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}
```

#### **Domain Model (Business Logic)**
```go
type User struct {
    ID           string     // Same
    Email        string     // Same
    Username     string     // Same
    FullName     string     // Same
    PhoneNumber  string     // Same
    PasswordHash string     // NOT in proto - Security!
    Status       UserStatus // Same
    CreatedAt    time.Time  // Different type (time.Time vs Timestamp)
    UpdatedAt    time.Time  // Different type
}
```

### **Key Differences & Why:**

1. **PasswordHash**: 
   - **Domain**: Has password hash untuk authentication
   - **Proto**: Tidak ada - security concern

2. **Data Types**:
   - **Domain**: `time.Time` (Go native)
   - **Proto**: `google.protobuf.Timestamp` (protobuf type)

3. **Additional Fields**:
   - **Domain**: Bisa ada fields yang tidak exposed via API
   - **Proto**: Only public API fields

4. **Business Methods**:
   - **Domain**: Has business logic methods
   - **Proto**: Only data, no behavior

## 7. **Conversion Between Layers**

### **Proto ↔ Domain Conversion**
```go
// In gRPC Handler
func (h *UserHandler) domainUserToProto(user *model.User) *pb.User {
    return &pb.User{
        Id:          user.ID,
        Email:       user.Email,
        Username:    user.Username,
        FullName:    user.FullName,
        PhoneNumber: user.PhoneNumber,
        Status:      h.domainStatusToProto(user.Status),
        CreatedAt:   timestamppb.New(user.CreatedAt),  // Type conversion
        UpdatedAt:   timestamppb.New(user.UpdatedAt),  // Type conversion
        // No PasswordHash - Security!
    }
}
```

### **Request DTO → Domain Conversion**
```go
// In Service Layer
func (r *CreateUserRequest) ToUser() *User {
    return &User{
        Email:       strings.ToLower(strings.TrimSpace(r.Email)),
        Username:    strings.ToLower(strings.TrimSpace(r.Username)),
        FullName:    strings.TrimSpace(r.FullName),
        PhoneNumber: strings.TrimSpace(r.PhoneNumber),
        Status:      UserStatusActive, // Business rule: default status
        CreatedAt:   time.Now(),       // Business rule: set timestamps
        UpdatedAt:   time.Now(),
        // PasswordHash will be set after hashing in service layer
    }
}
```

## 8. **Benefits of This Architecture**

### **1. Security**
- Password tidak pernah exposed di API responses
- Sensitive data terisolasi dalam domain model
- Clear separation antara internal dan external data

### **2. Flexibility**
- API bisa change tanpa affect domain model
- Domain model bisa evolve tanpa break API contract
- Easy untuk add new fields atau business rules

### **3. Testability**
- Each layer bisa di-test independently
- Business logic terisolasi dalam domain methods
- Easy untuk mock dependencies

### **4. Maintainability**
- Clear responsibility untuk each model type
- Easy untuk understand code purpose
- Consistent patterns across application

### **5. Performance**
- Only necessary data transferred over network
- Efficient database queries dengan proper mapping
- No unnecessary data serialization

## 9. **Real-World Example Flow**

### **Create User Flow:**
```
1. gRPC Request (Proto) → Handler
2. Handler converts Proto → CreateUserRequest DTO
3. Handler validates CreateUserRequest
4. Handler calls Service dengan DTO
5. Service converts DTO → Domain User
6. Service validates Domain User (business rules)
7. Service hashes password
8. Service calls Repository dengan Domain User
9. Repository saves Domain User ke database
10. Repository returns Domain User
11. Service returns Domain User
12. Handler converts Domain User → Proto Response
13. Handler returns Proto Response → Client
```

**Setiap step punya responsibility yang jelas dan terisolasi!**

## 10. **Conclusion**

Model yang kompleks ini **DIPERLUKAN** untuk:
- **Clean Architecture**: Proper separation of concerns
- **Security**: Protect sensitive data
- **Maintainability**: Easy untuk modify dan extend
- **Testability**: Each component testable independently
- **Business Logic**: Proper place untuk business rules
- **API Evolution**: API bisa change tanpa break internal logic

**Complexity is justified** karena memberikan **long-term benefits** dalam maintainability, security, dan scalability aplikasi enterprise-grade.
