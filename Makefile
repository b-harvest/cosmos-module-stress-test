#!/usr/bin/make -f

###############################################################################
###                                Build                                    ###
###############################################################################

install: go.sum
	@echo "Installing tester binary..."
	@go install ./cmd/tester

.PHONY: install

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################
test: test-unit

test-unit: 
	@go test -mod=readonly ./...

.PHONY: test test-unit

###############################################################################
###                               Localnet                                  ###
###############################################################################

localnet: 
	@echo "Bootstraping a single local testnet..."
	./scripts/localnet.sh

.PHONY: localnet