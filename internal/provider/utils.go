package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	bindings "github.com/srevinsaju/startrail-go-sdk"
)

func handleStartrailDiagnostics(diagnostics []bindings.Diagnostic, diags *diag.Diagnostics) {

	if diagnostics != nil {
		for _, d := range diagnostics {
			if d.Severity == "error" || d.Severity == "Error" {
				diags.AddError(d.Summary, d.Detail)
			} else if d.Severity == "warning" || d.Severity == "Warning" {
				diags.AddWarning(d.Summary, d.Detail)
			}
		}
	}
}
