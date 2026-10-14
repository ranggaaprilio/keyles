package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// defaultOrgName is used in invitation emails when the tenant name cannot be resolved
const defaultOrgName = "Keyles"

// buildInviteURL returns the frontend link for an invitation token.
// inviteBaseURL points at the frontend invite route, e.g. https://app.example.com/invite
func buildInviteURL(inviteBaseURL, plainToken string) string {
	return fmt.Sprintf("%s/%s", inviteBaseURL, plainToken)
}

// sendInvitationEmail sends the invitation email addressed from the tenant's organization.
// Delivery failures are logged rather than returned: the invitation is already persisted
// and the admin can resend it.
func sendInvitationEmail(
	ctx context.Context,
	emailService services.EmailService,
	tenantRepo repositories.TenantRepository,
	tenantID, toEmail, toName, inviteURL string,
) {
	orgName := defaultOrgName
	if id, err := uuid.Parse(tenantID); err == nil {
		if tenant, err := tenantRepo.FindByID(ctx, id); err == nil && tenant != nil && tenant.OrganizationName != "" {
			orgName = tenant.OrganizationName
		}
	}

	if err := emailService.SendInvitationEmail(ctx, toEmail, toName, inviteURL, orgName); err != nil {
		slog.WarnContext(ctx, "failed to send invitation email", "tenant_id", tenantID, "error", err)
	}
}
