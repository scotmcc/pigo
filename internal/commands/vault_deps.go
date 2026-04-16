package commands

import "github.com/scotmcc/pigo/internal/vault"

// vaultService is set at startup by SetVaultService.
// Vault commands check this before executing.
//
// Why this pattern: Commands register themselves at init() time (before main runs),
// but the vault.Service doesn't exist until config is loaded and dependencies are
// created. So commands register early with nil deps, and we wire them later.
// This is the Go equivalent of late-bound DI.
var vaultService *vault.Service

// SetVaultService wires the vault service into all vault commands.
// Called once at startup after the service is created.
func SetVaultService(svc *vault.Service) {
	vaultService = svc
}
