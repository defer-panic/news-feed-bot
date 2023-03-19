PROJECT_DIR = $(shell pwd)
PROJECT_BIN = $(PROJECT_DIR)/bin

MOQ = $(PROJECT_BIN)/moq
MOQ_VERSION = v0.3.1

.PHONY: .install-moq
.install-moq:
	@echo "Installing moq..."
	@mkdir -p $(PROJECT_BIN)
	[ -f $(MOQ) ] || go install github.com/matryer/moq@$(MOQ_VERSION)

.PHONY: install-env
install-env: .install-moq