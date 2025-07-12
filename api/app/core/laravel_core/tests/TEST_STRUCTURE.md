# Test Structure Documentation

## Overview

Our test structure is organized by the specific files/components being tested, making it easy to find and maintain tests as the codebase grows.

## Structure Principles

### Laravel Core Tests
- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test how components work together
- **Organization**: Grouped by the specific file/component being tested

### Go Core Tests  
- **Unit Tests**: Test individual core functions and methods
- **Integration Tests**: Test how core systems work together
- **Organization**: Grouped by the main component being tested

## Current Structure

```
api/app/core/
├── laravel_core/tests/
│   ├── unit/
│   │   └── facades/
│   │       └── config_facade_test.go          # Tests for config facade
│   ├── integration/
│   │   └── facades/
│   │       └── config_facade_integration_test.go  # Integration tests for config facade
│   └── run_tests.sh
└── go_core/tests/
    ├── unit/
    │   └── config/
    │       └── config_loader_test.go          # Tests for config loader
    ├── integration/
    │   └── config/
    │       └── config_system_test.go          # Integration tests for config system
    └── run_tests.sh
```

## Planned Structure

### Laravel Core Tests

#### Unit Tests (`api/app/core/laravel_core/tests/unit/`)
```
unit/
├── facades/
│   ├── config_facade_test.go
│   ├── cache_facade_test.go
│   ├── database_facade_test.go
│   ├── event_facade_test.go
│   ├── job_facade_test.go
│   ├── logging_facade_test.go
│   ├── mail_facade_test.go
│   └── service_facade_test.go
├── config/
│   └── config_test.go
├── http/
│   ├── base_controller_test.go
│   ├── base_request_test.go
│   └── requests/
│       ├── login_request_test.go
│       └── register_request_test.go
├── models/
│   ├── base_model_test.go
│   └── interfaces_test.go
├── providers/
│   ├── app_service_provider_test.go
│   ├── core_providers_test.go
│   ├── migration_service_provider_test.go
│   ├── service_provider_test.go
│   └── validation_service_provider_test.go
├── validators/
│   └── name_validator_test.go
├── listeners/
│   └── base_listener_test.go
├── observers/
│   └── model_observer_test.go
├── logging/
│   ├── logging_client_test.go
│   ├── logging_factory_test.go
│   ├── logging_provider_factory_test.go
│   ├── logging_providers_test.go
│   ├── logging_test.go
│   └── sentry_exception_handler_test.go
├── auth/
│   └── jwt_token_test.go
├── clients/
│   └── client_interfaces_test.go
├── dtos/
│   ├── base_dto_test.go
│   └── user_dto_test.go
├── env/
│   └── env_test.go
└── exceptions/
    ├── exceptions_test.go
    └── simple_exception_test.go
```

#### Integration Tests (`api/app/core/laravel_core/tests/integration/`)
```
integration/
├── facades/
│   ├── config_facade_integration_test.go
│   ├── cache_facade_integration_test.go
│   ├── database_facade_integration_test.go
│   ├── event_facade_integration_test.go
│   ├── job_facade_integration_test.go
│   ├── logging_facade_integration_test.go
│   ├── mail_facade_integration_test.go
│   └── service_facade_integration_test.go
├── http/
│   ├── controller_integration_test.go
│   ├── middleware_integration_test.go
│   └── request_integration_test.go
├── models/
│   └── model_integration_test.go
├── providers/
│   └── provider_integration_test.go
├── auth/
│   └── auth_integration_test.go
├── logging/
│   └── logging_integration_test.go
└── cross_component/
    ├── facade_to_core_integration_test.go
    ├── provider_to_facade_integration_test.go
    └── model_to_repository_integration_test.go
```

### Go Core Tests

#### Unit Tests (`api/app/core/go_core/tests/unit/`)
```
unit/
├── config/
│   └── config_loader_test.go
├── cache/
│   └── cache_test.go
├── container/
│   └── container_test.go
├── events/
│   └── events_test.go
├── mail/
│   └── mail_test.go
├── model/
│   └── model_test.go
├── queue/
│   └── queue_test.go
├── repository/
│   └── repository_test.go
├── validation/
│   └── validation_test.go
└── adapter/
    └── adapter_test.go
```

#### Integration Tests (`api/app/core/go_core/tests/integration/`)
```
integration/
├── config/
│   └── config_system_test.go
├── cache/
│   └── cache_system_test.go
├── container/
│   └── container_system_test.go
├── events/
│   └── events_system_test.go
├── mail/
│   └── mail_system_test.go
├── model/
│   └── model_system_test.go
├── queue/
│   └── queue_system_test.go
├── repository/
│   └── repository_system_test.go
├── validation/
│   └── validation_system_test.go
├── adapter/
│   └── adapter_system_test.go
└── cross_component/
    ├── config_to_cache_integration_test.go
    ├── events_to_queue_integration_test.go
    ├── model_to_repository_integration_test.go
    └── container_to_all_integration_test.go
```

## Naming Conventions

### File Names
- **Unit Tests**: `{component_name}_test.go`
- **Integration Tests**: `{component_name}_integration_test.go`
- **Cross-Component Tests**: `{component1}_to_{component2}_integration_test.go`

### Package Names
- **Unit Tests**: `package unit`
- **Integration Tests**: `package integration`

### Test Function Names
- **Unit Tests**: `Test{Component}{Method}` (e.g., `TestConfigLoaderLoad`)
- **Integration Tests**: `Test{Component}Integration` (e.g., `TestConfigFacadeIntegration`)

## Test Categories

### Unit Tests
- Test individual functions/methods in isolation
- Mock external dependencies
- Fast execution
- Focus on edge cases and error conditions

### Integration Tests
- Test how components work together
- Use real dependencies when possible
- Test complete workflows
- Focus on happy path and common scenarios

### Cross-Component Tests
- Test interactions between different core systems
- Ensure systems work together correctly
- Test performance and concurrency
- Validate architectural boundaries

## Running Tests

### Laravel Core Tests
```bash
# Run all Laravel core tests
cd api/app/core/laravel_core/tests
./run_tests.sh

# Run specific test categories
./run_tests.sh unit
./run_tests.sh integration
./run_tests.sh unit/facades
./run_tests.sh integration/facades
```

### Go Core Tests
```bash
# Run all Go core tests
cd api/app/core/go_core/tests
./run_tests.sh

# Run specific test categories
./run_tests.sh unit
./run_tests.sh integration
./run_tests.sh unit/config
./run_tests.sh integration/config
```

## Best Practices

1. **Test Organization**: Keep tests close to the code they test
2. **Clear Naming**: Use descriptive test and function names
3. **Isolation**: Each test should be independent and not rely on other tests
4. **Coverage**: Aim for high test coverage, especially for core functionality
5. **Performance**: Keep tests fast, especially unit tests
6. **Maintainability**: Write tests that are easy to understand and maintain
7. **Documentation**: Document complex test scenarios and edge cases

## Adding New Tests

When adding new tests:

1. **Identify the component** being tested
2. **Choose the appropriate category** (unit vs integration)
3. **Create the folder structure** if it doesn't exist
4. **Follow naming conventions** for files and functions
5. **Update this documentation** if adding new test categories
6. **Update test runners** if adding new test directories 