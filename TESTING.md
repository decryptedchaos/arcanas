# Unit Testing Guide for Arcanas

## What are Unit Tests?

Unit tests are automated tests that verify individual functions or components work correctly in isolation. They help catch bugs early, enable confident refactoring, and serve as living documentation.

## Benefits

- ✅ **Catch bugs early** - Find issues before production
- ✅ **Refactoring confidence** - Change code without fear
- ✅ **Documentation** - Tests show how code should be used
- ✅ **Faster debugging** - Pinpoint exactly what broke
- ✅ **Better design** - Testable code is better code

## Go Testing Basics

### File Structure
- Test files end with `_test.go`
- Place test files next to the code they test
- Example: `pools.go` → `pools_test.go`

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./internal/system

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run a specific test
go test -run TestGetStoragePools ./internal/system
```

## Example Test Structure

### Basic Test Example

```go
package system

import (
    "testing"
)

func TestGetStoragePools(t *testing.T) {
    // Arrange - Set up test data
    
    // Act - Execute the function
    pools, err := GetStoragePools()
    
    // Assert - Verify the results
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    if pools == nil {
        t.Error("Expected pools to not be nil")
    }
}
```

### Table-Driven Tests (Recommended)

```go
func TestCreateStoragePool(t *testing.T) {
    tests := []struct {
        name    string
        input   models.StoragePoolCreateRequest
        wantErr bool
    }{
        {
            name: "valid mergerfs pool",
            input: models.StoragePoolCreateRequest{
                Name:    "test-pool",
                Type:    "mergerfs",
                Devices: []string{"/dev/sda1"},
            },
            wantErr: false,
        },
        {
            name: "missing name",
            input: models.StoragePoolCreateRequest{
                Type:    "mergerfs",
                Devices: []string{"/dev/sda1"},
            },
            wantErr: true,
        },
        {
            name: "no devices",
            input: models.StoragePoolCreateRequest{
                Name: "test-pool",
                Type: "mergerfs",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CreateStoragePool(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateStoragePool() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Testing Best Practices

### 1. Use Descriptive Test Names
```go
// Good
func TestCreateStoragePool_WithInvalidName_ReturnsError(t *testing.T) {}

// Bad
func TestPool(t *testing.T) {}
```

### 2. Test One Thing Per Test
```go
// Good - Focused test
func TestGetStoragePools_ReturnsNonEmptyList(t *testing.T) {}
func TestGetStoragePools_HandlesNoPoolsGracefully(t *testing.T) {}

// Bad - Testing too much
func TestGetStoragePools(t *testing.T) {
    // Tests 5 different scenarios
}
```

### 3. Use Test Helpers
```go
func createTestPool(t *testing.T, name string) models.StoragePool {
    t.Helper() // Marks this as a helper function
    pool := models.StoragePool{Name: name}
    // ... setup code
    return pool
}
```

### 4. Clean Up After Tests
```go
func TestCreateStoragePool(t *testing.T) {
    // Create test resources
    
    // Clean up when test completes
    t.Cleanup(func() {
        // Delete test pool
    })
}
```

## Mocking System Commands

Since Arcanas interacts with system commands, you'll need to mock them for testing:

```go
// Create an interface for command execution
type CommandExecutor interface {
    Execute(cmd string, args ...string) (string, error)
}

// Real implementation
type RealExecutor struct{}

func (r *RealExecutor) Execute(cmd string, args ...string) (string, error) {
    out, err := exec.Command(cmd, args...).Output()
    return string(out), err
}

// Mock implementation for testing
type MockExecutor struct {
    Output string
    Err    error
}

func (m *MockExecutor) Execute(cmd string, args ...string) (string, error) {
    return m.Output, m.Err
}

// Use in tests
func TestGetStoragePools_WithMock(t *testing.T) {
    mockExec := &MockExecutor{
        Output: "pool1\npool2\n",
        Err:    nil,
    }
    
    // Use mockExec instead of real commands
}
```

## Recommended Testing Libraries

### testify/assert
Makes assertions more readable:

```bash
go get github.com/stretchr/testify/assert
```

```go
import "github.com/stretchr/testify/assert"

func TestGetStoragePools(t *testing.T) {
    pools, err := GetStoragePools()
    
    assert.NoError(t, err)
    assert.NotNil(t, pools)
    assert.Greater(t, len(pools), 0)
}
```

### testify/mock
For creating mocks:

```go
import "github.com/stretchr/testify/mock"

type MockStorage struct {
    mock.Mock
}

func (m *MockStorage) GetPools() ([]Pool, error) {
    args := m.Called()
    return args.Get(0).([]Pool), args.Error(1)
}
```

## Coverage Goals

- **Aim for 70-80% coverage** for critical business logic
- **100% coverage** for utility functions and helpers
- **Focus on critical paths** rather than chasing 100% everywhere

## Next Steps

1. Start with testing pure functions (no I/O)
2. Add tests for business logic in `internal/system/`
3. Mock system commands for integration tests
4. Add handler tests with mock HTTP requests
5. Set up CI/CD to run tests automatically

## Example: First Test to Write

Start with testing the storage pool validation logic:

```go
// internal/system/pools_test.go
package system

import (
    "testing"
    "arcanas/internal/models"
)

func TestValidateStoragePoolRequest(t *testing.T) {
    tests := []struct {
        name    string
        req     models.StoragePoolCreateRequest
        wantErr bool
    }{
        {
            name: "valid request",
            req: models.StoragePoolCreateRequest{
                Name:    "my-pool",
                Type:    "mergerfs",
                Devices: []string{"/dev/sda1"},
            },
            wantErr: false,
        },
        {
            name: "empty name",
            req: models.StoragePoolCreateRequest{
                Type:    "mergerfs",
                Devices: []string{"/dev/sda1"},
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validatePoolRequest(tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```
