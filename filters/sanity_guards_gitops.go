package filters

import "strings"

// looksLike* guards for kustomize, argocd, flux, and stern filters.

func looksLikeKustomizeBuildOutput(s string) bool {
	return strings.Contains(s, "apiVersion:") ||
		strings.Contains(s, "kind:") ||
		strings.Contains(s, "---")
}

func looksLikeArgoCDAppListOutput(s string) bool {
	return strings.Contains(s, "STATUS") &&
		strings.Contains(s, "HEALTH") &&
		(strings.Contains(s, "Synced") || strings.Contains(s, "OutOfSync") ||
			strings.Contains(s, "Healthy") || strings.Contains(s, "Degraded") ||
			strings.Contains(s, "NAME"))
}

func looksLikeArgoCDSyncOutput(s string) bool {
	return (strings.Contains(s, "TIMESTAMP") && strings.Contains(s, "KIND")) ||
		(strings.Contains(s, "Sync Status:") || strings.Contains(s, "Health Status:"))
}

func looksLikeArgoCDAppGetOutput(s string) bool {
	return (strings.Contains(s, "Sync Status:") || strings.Contains(s, "Health Status:")) &&
		strings.Contains(s, "Name:")
}

func looksLikeFluxGetOutput(s string) bool {
	return strings.Contains(s, "READY") &&
		strings.Contains(s, "MESSAGE") &&
		(strings.Contains(s, "Applied revision") ||
			strings.Contains(s, "True") ||
			strings.Contains(s, "False") ||
			strings.Contains(s, "NAME"))
}

func looksLikeFluxReconcileOutput(s string) bool {
	return strings.Contains(s, "annotating") ||
		strings.Contains(s, "reconciliation") ||
		strings.Contains(s, "fetched revision") ||
		strings.Contains(s, "applied revision") ||
		strings.Contains(s, "✔") ||
		strings.Contains(s, "✗") ||
		strings.Contains(s, "►") ||
		strings.Contains(s, "◎")
}

func looksLikeSternOutput(_ string) bool {
	// stern output can be anything (multi-pod logs) — always attempt
	return true
}
