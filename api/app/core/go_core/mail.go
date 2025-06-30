package go_core

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

// Email represents a generic email message
type Email[T any] struct {
	ID          string            `json:"id"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Cc          []string          `json:"cc,omitempty"`
	Bcc         []string          `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	HTMLBody    string            `json:"html_body,omitempty"`
	Attachments []*Attachment     `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Data        T                 `json:"data,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	SentAt      *time.Time        `json:"sent_at,omitempty"`
	Status      EmailStatus       `json:"status"`
	Error       string            `json:"error,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Name        string `json:"name"`
	Content     []byte `json:"content"`
	ContentType string `json:"content_type"`
	Inline      bool   `json:"inline"`
}

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending EmailStatus = "pending"
	EmailStatusSent    EmailStatus = "sent"
	EmailStatusFailed  EmailStatus = "failed"
)

// Mailer defines a generic mailer interface
type Mailer[T any] interface {
	// Basic operations
	Send(email *Email[T]) error
	SendAsync(email *Email[T]) error
	SendMany(emails []*Email[T]) error

	// Template operations
	SendTemplate(template string, data T, to []string) error
	SendTemplateAsync(template string, data T, to []string) error

	// Utility operations
	Validate(email *Email[T]) error
	WithContext(ctx context.Context) Mailer[T]
}

// MailTemplate defines a template interface
type MailTemplate[T any] interface {
	// Template operations
	Render(templateName string, data T) (*Email[T], error)
	RenderHTML(templateName string, data T) (string, error)
	RenderText(templateName string, data T) (string, error)

	// Template management
	AddTemplate(name string, template string) error
	RemoveTemplate(name string) error
	HasTemplate(name string) bool
}

// smtpMailer implements Mailer[T] with SMTP
type smtpMailer[T any] struct {
	host     string
	port     int
	username string
	password string
	from     string
	ctx      context.Context
}

// NewSMTPMailer creates a new SMTP mailer instance
func NewSMTPMailer[T any](host string, port int, username, password, from string) Mailer[T] {
	return &smtpMailer[T]{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		ctx:      context.Background(),
	}
}

// Send sends an email synchronously
func (m *smtpMailer[T]) Send(email *Email[T]) error {
	// Validate email
	err := m.Validate(email)
	if err != nil {
		return err
	}

	// Set default from address if not provided
	if email.From == "" {
		email.From = m.from
	}

	// Build message
	message, err := m.buildMessage(email)
	if err != nil {
		return err
	}

	// Send email
	err = m.sendMessage(email.To, message)
	if err != nil {
		email.Status = EmailStatusFailed
		email.Error = err.Error()
		return err
	}

	// Update status
	now := time.Now()
	email.SentAt = &now
	email.Status = EmailStatusSent

	return nil
}

// SendAsync sends an email asynchronously
func (m *smtpMailer[T]) SendAsync(email *Email[T]) error {
	go func() {
		_ = m.Send(email)
	}()
	return nil
}

// SendMany sends multiple emails
func (m *smtpMailer[T]) SendMany(emails []*Email[T]) error {
	for _, email := range emails {
		err := m.Send(email)
		if err != nil {
			return err
		}
	}
	return nil
}

// SendTemplate sends an email using a template
func (m *smtpMailer[T]) SendTemplate(template string, data T, to []string) error {
	// This would require a template engine
	// For now, we'll create a basic email
	email := &Email[T]{
		To:        to,
		Subject:   "Template Email",
		Body:      "Template content",
		Data:      data,
		CreatedAt: time.Now(),
		Status:    EmailStatusPending,
	}

	return m.Send(email)
}

// SendTemplateAsync sends an email using a template asynchronously
func (m *smtpMailer[T]) SendTemplateAsync(template string, data T, to []string) error {
	go func() {
		_ = m.SendTemplate(template, data, to)
	}()
	return nil
}

// Validate validates an email
func (m *smtpMailer[T]) Validate(email *Email[T]) error {
	if email.To == nil || len(email.To) == 0 {
		return fmt.Errorf("recipient list is empty")
	}

	if email.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if email.Body == "" && email.HTMLBody == "" {
		return fmt.Errorf("email body is required")
	}

	return nil
}

// WithContext returns a mailer with context
func (m *smtpMailer[T]) WithContext(ctx context.Context) Mailer[T] {
	return &smtpMailer[T]{
		host:     m.host,
		port:     m.port,
		username: m.username,
		password: m.password,
		from:     m.from,
		ctx:      ctx,
	}
}

// buildMessage builds the email message
func (m *smtpMailer[T]) buildMessage(email *Email[T]) ([]byte, error) {
	var message strings.Builder

	// Add headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", email.From))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))

	if len(email.Cc) > 0 {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.Cc, ", ")))
	}

	message.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	message.WriteString("MIME-Version: 1.0\r\n")

	// Add custom headers
	for key, value := range email.Headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Handle attachments
	if len(email.Attachments) > 0 {
		boundary := "boundary123"
		message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

		// Add text body
		if email.Body != "" {
			message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			message.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
			message.WriteString(email.Body)
			message.WriteString("\r\n\r\n")
		}

		// Add HTML body
		if email.HTMLBody != "" {
			message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			message.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
			message.WriteString(email.HTMLBody)
			message.WriteString("\r\n\r\n")
		}

		// Add attachments
		for _, attachment := range email.Attachments {
			message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			message.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
			message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name))
			message.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
			message.WriteString(fmt.Sprintf("%s\r\n", attachment.Content))
		}

		message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Simple email without attachments
		if email.HTMLBody != "" {
			message.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
			message.WriteString(email.HTMLBody)
		} else {
			message.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
			message.WriteString(email.Body)
		}
	}

	return []byte(message.String()), nil
}

// sendMessage sends the actual message via SMTP
func (m *smtpMailer[T]) sendMessage(to []string, message []byte) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	// Create auth
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	// Send email
	return smtp.SendMail(addr, auth, m.from, to, message)
}

// localMailer implements Mailer[T] for local development
type localMailer[T any] struct {
	emails []*Email[T]
	ctx    context.Context
}

// NewLocalMailer creates a new local mailer for development
func NewLocalMailer[T any]() Mailer[T] {
	return &localMailer[T]{
		emails: make([]*Email[T], 0),
		ctx:    context.Background(),
	}
}

// Send stores the email locally (for development)
func (m *localMailer[T]) Send(email *Email[T]) error {
	// Validate email
	err := m.Validate(email)
	if err != nil {
		return err
	}

	// Store email locally
	m.emails = append(m.emails, email)

	// Update status
	now := time.Now()
	email.SentAt = &now
	email.Status = EmailStatusSent

	return nil
}

// SendAsync stores the email locally asynchronously
func (m *localMailer[T]) SendAsync(email *Email[T]) error {
	go func() {
		_ = m.Send(email)
	}()
	return nil
}

// SendMany stores multiple emails locally
func (m *localMailer[T]) SendMany(emails []*Email[T]) error {
	for _, email := range emails {
		err := m.Send(email)
		if err != nil {
			return err
		}
	}
	return nil
}

// SendTemplate stores a template email locally
func (m *localMailer[T]) SendTemplate(template string, data T, to []string) error {
	email := &Email[T]{
		To:        to,
		Subject:   fmt.Sprintf("[TEMPLATE] %s", template),
		Body:      "Template content",
		Data:      data,
		CreatedAt: time.Now(),
		Status:    EmailStatusPending,
	}

	return m.Send(email)
}

// SendTemplateAsync stores a template email locally asynchronously
func (m *localMailer[T]) SendTemplateAsync(template string, data T, to []string) error {
	go func() {
		_ = m.SendTemplate(template, data, to)
	}()
	return nil
}

// Validate validates an email
func (m *localMailer[T]) Validate(email *Email[T]) error {
	if email.To == nil || len(email.To) == 0 {
		return fmt.Errorf("recipient list is empty")
	}

	if email.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if email.Body == "" && email.HTMLBody == "" {
		return fmt.Errorf("email body is required")
	}

	return nil
}

// WithContext returns a mailer with context
func (m *localMailer[T]) WithContext(ctx context.Context) Mailer[T] {
	return &localMailer[T]{
		emails: m.emails,
		ctx:    ctx,
	}
}

// GetEmails returns all stored emails (for development)
func (m *localMailer[T]) GetEmails() []*Email[T] {
	return m.emails
}

// ClearEmails clears all stored emails (for development)
func (m *localMailer[T]) ClearEmails() {
	m.emails = make([]*Email[T], 0)
}

// MailStore defines an interface for persisting emails
type MailStore[T any] interface {
	// Email persistence
	Store(email *Email[T]) error
	StoreMany(emails []*Email[T]) error

	// Email retrieval
	Get(emailID string) (*Email[T], error)
	GetByStatus(status EmailStatus) ([]*Email[T], error)
	GetByTimeRange(start, end time.Time) ([]*Email[T], error)

	// Email management
	Update(email *Email[T]) error
	Delete(emailID string) error
	Clear() error

	// Utility operations
	Count() (int64, error)
	CountByStatus(status EmailStatus) (int64, error)
}

// memoryMailStore implements MailStore[T] with in-memory storage
type memoryMailStore[T any] struct {
	emails map[string]*Email[T]
	mu     sync.RWMutex
}

// NewMemoryMailStore creates a new in-memory mail store
func NewMemoryMailStore[T any]() MailStore[T] {
	return &memoryMailStore[T]{
		emails: make(map[string]*Email[T]),
	}
}

// Store stores an email
func (s *memoryMailStore[T]) Store(email *Email[T]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.emails[email.ID] = email
	return nil
}

// StoreMany stores multiple emails
func (s *memoryMailStore[T]) StoreMany(emails []*Email[T]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, email := range emails {
		s.emails[email.ID] = email
	}

	return nil
}

// Get retrieves an email by ID
func (s *memoryMailStore[T]) Get(emailID string) (*Email[T], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	email, exists := s.emails[emailID]
	if !exists {
		return nil, fmt.Errorf("email not found: %s", emailID)
	}

	return email, nil
}

// GetByStatus retrieves emails by status
func (s *memoryMailStore[T]) GetByStatus(status EmailStatus) ([]*Email[T], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var emails []*Email[T]
	for _, email := range s.emails {
		if email.Status == status {
			emails = append(emails, email)
		}
	}

	return emails, nil
}

// GetByTimeRange retrieves emails within a time range
func (s *memoryMailStore[T]) GetByTimeRange(start, end time.Time) ([]*Email[T], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var emails []*Email[T]
	for _, email := range s.emails {
		if email.CreatedAt.After(start) && email.CreatedAt.Before(end) {
			emails = append(emails, email)
		}
	}

	return emails, nil
}

// Update updates an email
func (s *memoryMailStore[T]) Update(email *Email[T]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.emails[email.ID] = email
	return nil
}

// Delete removes an email
func (s *memoryMailStore[T]) Delete(emailID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.emails, emailID)
	return nil
}

// Clear removes all emails
func (s *memoryMailStore[T]) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.emails = make(map[string]*Email[T])
	return nil
}

// Count returns the total number of emails
func (s *memoryMailStore[T]) Count() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.emails)), nil
}

// CountByStatus returns the number of emails by status
func (s *memoryMailStore[T]) CountByStatus(status EmailStatus) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := int64(0)
	for _, email := range s.emails {
		if email.Status == status {
			count++
		}
	}

	return count, nil
}
