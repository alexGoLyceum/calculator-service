AGENT=cmd/agent/main.go
ORCHESTRATOR=cmd/orchestrator/main.go
BUILD_DIR=bin
AGENT_BIN=$(BUILD_DIR)/agent
ORCHESTRATOR_BIN=$(BUILD_DIR)/orchestrator

.PHONY: build
build:
	@echo "Building binaries..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(AGENT_BIN) $(AGENT)
	@go build -o $(ORCHESTRATOR_BIN) $(ORCHESTRATOR)

.PHONY: run
run: build
	@echo "Running agent and orchestrator..."
	@$(ORCHESTRATOR_BIN) & echo $$! > $(BUILD_DIR)/orchestrator.pid
	@sleep 2
	@$(AGENT_BIN) & echo $$! > $(BUILD_DIR)/agent.pid

.PHONY: stop
stop:
	@echo "Stopping processes..."
	@[ -f $(BUILD_DIR)/agent.pid ] && kill -9 `cat $(BUILD_DIR)/agent.pid` && rm $(BUILD_DIR)/agent.pid || true
	@[ -f $(BUILD_DIR)/orchestrator.pid ] && kill -9 `cat $(BUILD_DIR)/orchestrator.pid` && rm $(BUILD_DIR)/orchestrator.pid || true

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

.PHONY: rebuild
rebuild: clean build

.PHONY: help
help:
	@echo "Available make commands:"
	@echo "  build        - Compile the binaries"
	@echo "  run          - Run agent and orchestrator"
	@echo "  stop         - Stop running processes"
	@echo "  clean        - Remove built binaries"
	@echo "  rebuild      - Clean and rebuild everything"
	@echo "  help         - Show this help message"