# TDD Development Plan

## TDD Principles
1. **Red**: Write a failing test first
2. **Green**: Write minimal code to make the test pass
3. **Refactor**: Improve code while keeping tests green

## Project Structure (Test-First)

```
gxf-discord-bot/
├── pkg/                    # Public packages
│   ├── config/            # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── bot/               # Discord bot core
│   │   ├── bot.go
│   │   └── bot_test.go
│   ├── action/            # Action handlers
│   │   ├── action.go
│   │   ├── action_test.go
│   │   ├── command.go
│   │   ├── command_test.go
│   │   ├── message.go
│   │   ├── message_test.go
│   │   ├── reaction.go
│   │   ├── reaction_test.go
│   │   ├── scheduler.go
│   │   └── scheduler_test.go
│   ├── response/          # Response handlers
│   │   ├── response.go
│   │   └── response_test.go
│   ├── auth/              # OAuth authentication
│   │   ├── auth.go
│   │   └── auth_test.go
│   ├── secrets/           # Secret management
│   │   ├── secrets.go
│   │   ├── secrets_test.go
│   │   ├── vault.go
│   │   └── vault_test.go
│   └── ratelimit/         # Rate limiting
│       ├── ratelimit.go
│       └── ratelimit_test.go
├── cmd/                   # CLI commands
│   ├── root.go
│   ├── root_test.go
│   ├── generate.go
│   ├── generate_test.go
│   ├── validate.go
│   └── validate_test.go
├── internal/              # Private packages
│   └── testutil/         # Test utilities & mocks
│       ├── mocks.go
│       └── fixtures.go
├── test/                  # Integration tests
│   ├── integration/
│   │   └── bot_test.go
│   └── e2e/
│       └── bot_e2e_test.go
├── examples/              # Example configurations
│   ├── basic/
│   ├── oauth/
│   └── vault/
├── schema/                # CUE schemas
│   └── config.cue
└── main.go               # Entry point
```

## Development Order (TDD Cycles)

### Phase 1: Configuration (Week 1)
1. Write test for loading YAML config
2. Write test for config validation
3. Write test for environment variable substitution
4. Write test for CUE schema validation

### Phase 2: Core Bot (Week 1-2)
1. Write test for bot initialization
2. Write test for Discord connection
3. Write test for event handler registration
4. Write test for graceful shutdown

### Phase 3: Actions (Week 2-3)
1. Write tests for command parsing
2. Write tests for pattern matching
3. Write tests for reaction handling
4. Write tests for scheduled tasks

### Phase 4: Responses (Week 3)
1. Write tests for text responses
2. Write tests for embed responses
3. Write tests for DM responses
4. Write tests for HTTP webhooks

### Phase 5: Auth & Security (Week 4)
1. Write tests for OAuth flow
2. Write tests for user authorization
3. Write tests for role-based access
4. Write tests for rate limiting

### Phase 6: Secrets (Week 4-5)
1. Write tests for Vault connection
2. Write tests for token retrieval
3. Write tests for auth methods (token, approle, k8s)

### Phase 7: Integration (Week 5)
1. Write integration tests
2. Write E2E tests
3. Add example configurations

## Test Coverage Goals
- **Target**: 100% coverage for all packages
- **Minimum**: 90% coverage
- Run coverage reports after each phase

## Testing Tools
- `testing`: Go standard library
- `testify/assert`: Assertions
- `testify/mock`: Mocking
- `testify/suite`: Test suites
- `go-cmp`: Deep comparisons

## Continuous Testing
```bash
# Watch mode for TDD
make test-watch

# Coverage report
make test-coverage

# Race detector
make test-race
```
