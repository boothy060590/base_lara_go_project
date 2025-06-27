# Email Templating System

This directory contains the email templates for the Base Laravel Go Project, following Laravel-style patterns.

## Directory Structure

```
views/mail/
├── base.html                    # Base template with common styling and layout
├── auth/                        # Authentication-related email templates
│   ├── welcome.html            # Welcome email for new users
│   ├── password_reset.html     # Password reset email
│   └── email_verification.html # Email verification email
└── README.md                   # This documentation
```

## Template Structure

### Base Template (`base.html`)

The base template provides:
- Consistent styling and layout
- Responsive design
- Common header and footer
- CSS-in-HTML for email client compatibility

### Content Templates

Each content template defines a `content` block that gets inserted into the base template:

```html
{{define "content"}}
<div class="content-section">
    <h2>Welcome to {{.AppName}}, {{.User.FirstName}}!</h2>
    <!-- Template content here -->
</div>
{{end}}
```

## Available Variables

All templates have access to these variables:

- `{{.Subject}}` - Email subject
- `{{.AppName}}` - Application name
- `{{.Year}}` - Current year
- `{{.RecipientEmail}}` - Recipient's email address
- `{{.User}}` - User object with properties like FirstName, LastName, Email, etc.
- `{{.LoginURL}}` - Login page URL
- `{{.ResetURL}}` - Password reset URL (for password reset emails)
- `{{.VerificationURL}}` - Email verification URL (for verification emails)

## Usage Examples

### Sending a Welcome Email

```go
import (
    "base_lara_go_project/app/core"
    "base_lara_go_project/app/facades"
)

// Prepare template data
templateData := core.EmailTemplateData{
    Subject:        "Welcome to Our App!",
    AppName:        "My Application",
    RecipientEmail: user.Email,
    User:           user,
    LoginURL:       "https://app.example.com/login",
}

// Send email using template
err := facades.MailTemplateToUser(user, "auth/welcome", templateData)
```

### Sending a Password Reset Email

```go
templateData := core.EmailTemplateData{
    Subject:        "Password Reset Request",
    AppName:        "My Application",
    RecipientEmail: user.Email,
    User:           user,
    ResetURL:       "https://app.example.com/reset-password?token=" + resetToken,
}

err := facades.MailTemplateToUser(user, "auth/password_reset", templateData)
```

## Creating New Templates

1. Create a new HTML file in the appropriate subdirectory
2. Define the `content` block with your template content
3. Use the available variables for dynamic content
4. Test the template using the test endpoint

### Example: Creating a Notification Template

```html
<!-- views/mail/notifications/account_updated.html -->
{{define "content"}}
<div class="content-section">
    <h2>Account Updated</h2>
    
    <p>Hello {{.User.FirstName}},</p>
    
    <p>Your account information has been successfully updated.</p>
    
    <div class="highlight-box">
        <h3>Updated Information</h3>
        <p><strong>Updated at:</strong> {{.UpdatedAt}}</p>
        <p><strong>Updated fields:</strong> {{.UpdatedFields}}</p>
    </div>
    
    <p>If you didn't make these changes, please contact our support team immediately.</p>
    
    <p>Best regards,<br>The {{.AppName}} Team</p>
</div>
{{end}}
```

## Styling Guidelines

- Use inline CSS for maximum email client compatibility
- Keep styles simple and avoid complex layouts
- Test in multiple email clients
- Use the provided CSS classes for consistency:
  - `.highlight-box` - For important information
  - `.info-box` - For informational content
  - `.btn` - For call-to-action buttons
  - `.content-section` - For content sections

## Testing

Use the test endpoint to verify your templates:

```bash
curl -X POST https://api.baselaragoproject.test/v1/auth/test-email-template
```

This will send a test welcome email to `john.doe@example.com` using the `auth/welcome` template.

## Best Practices

1. **Keep templates simple** - Email clients have limited CSS support
2. **Use semantic HTML** - Helps with accessibility and email client rendering
3. **Test thoroughly** - Test in multiple email clients (Gmail, Outlook, Apple Mail, etc.)
4. **Use the base template** - Ensures consistency across all emails
5. **Validate variables** - Always check if variables exist before using them
6. **Keep file sizes small** - Large emails may be blocked by spam filters

## Troubleshooting

### Template Not Found Error
- Ensure the template file exists in the correct directory
- Check that the template name matches the file path (without .html extension)
- Verify the template is being loaded during application startup

### Rendering Issues
- Check that all required variables are provided
- Ensure the template syntax is correct
- Verify that the base template is accessible

### Email Not Sending
- Check the mail configuration in your environment
- Verify that the mail service is properly initialized
- Check the application logs for error messages 