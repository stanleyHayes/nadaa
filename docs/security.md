# Security

## MVP Controls

- Role-based access control for authority workflows.
- MFA for authority users.
- Audit logs for alert, incident, assignment, status, and admin actions.
- Explicit contact permission for citizen reports.
- Anonymous report support where policy allows.
- Rate limits for public incident intake.
- Private media storage with controlled access.
- Approval workflow for mass alerts.
- Emergency override restricted to authorized roles and fully audited.

## Secret Handling

Never commit:

- API tokens.
- Client or citizen data.
- Database credentials.
- GitHub secrets.
- Jira credentials.
- SMS, WhatsApp, push, email, or cloud provider credentials.

Use environment variables and deployment secret stores.

## AI/ML Safety

- ML predictions must not automatically issue public alerts.
- Model outputs must show confidence and explanation factors.
- Model version and prediction inputs must be logged.
- False positives and false negatives must be reviewable.

